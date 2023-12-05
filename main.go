package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"pdfwtk/pkg"
	"strconv"

	"github.com/pdfcpu/pdfcpu/pkg/api"
	"github.com/pdfcpu/pdfcpu/pkg/pdfcpu/model"
	"github.com/tidwall/gjson"
)

var rpc *bool

func init() {
	rpc = flag.Bool("rpc", false, "use rpc")
	flag.Parse()
}

// func index(w http.ResponseWriter, r *http.Request) {
// 	// r.ParseForm()
// 	// fmt.Println("Form: ", r.Form)
// 	// fmt.Println("Path: ", r.URL.Path)
// 	// for k, v := range r.Form {
// 	// 	fmt.Println(k, "=>", v, strings.Join(v, "-"))
// 	// }
// 	fmt.Fprint(w, "It works !")
// }

func pdfmerge(w http.ResponseWriter, r *http.Request) {

	resp := "Success"
	body, _ := io.ReadAll(r.Body)
	//output file name
	outfile := gjson.GetBytes(body, "outfile").String()
	// fmt.Println(outfile)

	// input file names
	var infiles []string
	result := []byte(gjson.GetBytes(body, "infiles").Raw)
	// fmt.Println(string(result))
	json.Unmarshal(result, &infiles)

	if len(infiles) <= 0 {
		resp = "Error: no infiles"
	} else {
		log.Printf("%s ready to merge...", outfile)
		// configureation
		conf := model.NewDefaultConfiguration()

		// don't validate
		conf.ValidationMode = model.ValidationNone
		conf.CreateBookmarks = false
		conf.OptimizeDuplicateContentStreams = true
		// conf.WriteObjectStream = false
		// conf.WriteXRefStream = false
		// conf.HeaderBufSize = 100

		// fmt.Println(conf.ValidationModeString())

		// encrypt check and permission set
		enc := gjson.GetBytes(body, "protect").Int() == 1
		if enc {
			// fmt.Println("need protect!")
			pwd := gjson.GetBytes(body, "password").String()
			conf.OwnerPW = pwd
			conf.UserPW = ""
			conf.EncryptUsingAES = true
			conf.EncryptKeyLength = 256
			conf.Permissions = 2252
			// 204 + 2048
			if len(pwd) > 0 {
				log.Println("\t|... need protect!")
			}
		}

		//validate
		allfile := 0
		for _, v := range infiles {
			// fmt.Println(v)
			finfo, err := os.Stat(v)
			if err != nil {
				log.Printf("\t|... %s not found!", v)
				// return
			} else {
				if finfo.IsDir() {
					log.Printf("\t|... %s is directory!", v)
				} else {
					allfile += 1
				}
			}
		}

		// Create new Merged or/and Encrypt pdf
		if len(infiles) != allfile {
			log.Printf("\t|... error: %v infiles, %v validated! Merge failed: %s", len(infiles), allfile, outfile)
			// remove outfile if exists
			ofinfo, err := os.Stat(outfile)
			if err == nil {
				if !ofinfo.IsDir() {
					os.Remove(outfile)
					// log.Printf("\t|... delete %s!", outfile)
				}
			}
			resp = "Error: " + strconv.Itoa(len(infiles)) + " infiles, but only " + strconv.Itoa(allfile) + " validated!"
		} else {

			if !*rpc {
				// local
				err := pkg.MergeCreateFile(infiles, outfile, conf)
				if err != nil {
					resp = "Error for Merge: " + string(err.Error())
					log.Printf("\t|... %s", resp)
				}
				if enc && (conf.OwnerPW != "" || conf.UserPW != "") {
					err = api.EncryptFile(outfile, "", conf)
					if err != nil {
						resp = "Error for Encrypt: " + string(err.Error())
						log.Printf("\t|... %s", resp)
					}
				}
			} else {

				// CallPpc
				reply, err := CallRpc(infiles, outfile, conf)
				if err != nil {
					resp = "Error for Merge: " + string(err.Error())
					log.Printf("\t|... %s", resp)
				} else {
					// if enc && (conf.OwnerPW != "" || conf.UserPW != "") {
					// 	err = api.EncryptFile(outfile, "", conf)
					// 	if err != nil {
					// 		resp = "Error for Encrypt: " + string(err.Error())
					// 		log.Printf("\t|... %s", resp)
					// 	}
					// }
					w, _ := os.Create(outfile)
					// if err != nil {
					// 	log.Fatal(err)
					// }

					defer func() {
						// if err = w.Close(); err != nil {
						// 	return
						// }
						w.Close()
					}()
					w.Write(reply.W)
				}
			}

			if err := os.Chmod(outfile, 0666); err != nil {
				log.Printf("\t|... error: %s", err)
			}
			// u, err := user.Lookup("www-data")
			// if err != nil {
			// 	log.Println("no user www-data")
			// } else {
			// 	uid, _ := strconv.Atoi(u.Uid)
			// 	gid, _ := strconv.Atoi(u.Gid)
			// 	if err := os.Chown(outfile, uid, gid); err != nil {
			// 		log.Printf("\t|... error: %s", err)
			// 	}
			// }
		}
	}

	log.Printf("Merge to %s %s!\n", outfile, resp)
	fmt.Fprint(w, resp)
}

func pdfvalidate(w http.ResponseWriter, r *http.Request) {

	resp := "Success"
	body, _ := io.ReadAll(r.Body)

	// input file names
	var infiles []string
	result := []byte(gjson.GetBytes(body, "infiles").Raw)
	// fmt.Println(string(result))
	json.Unmarshal(result, &infiles)

	if len(infiles) <= 0 {
		resp = "Error: no infiles"
	} else {
		// configureation
		conf := model.NewDefaultConfiguration()

		// validate mode
		// conf.ValidationMode = model.ValidationRelaxed
		// fmt.Println(conf.ValidationModeString())

		//validate
		if !*rpc {
			// local
			for _, fn := range infiles {
				err := api.ValidateFile(fn, conf)
				if err != nil {
					if resp == "Success" {
						resp = "Error: \n"
					}
					resp += fmt.Sprintf("validate error: %s, %s\n", fn, err)
					log.Printf("\t|... %s, %s\n", fn, resp)
				}
			}
			// err := api.ValidateFiles(infiles, conf)
			// if err != nil {
			// 	resp = "validate error: " + string(err.Error())
			// 	log.Printf("\t|... %s", resp)
			// }

			// } else {

			// 	// CallPpc
			// 	reply, err := CallRpc(infiles, outfile, conf)
			// 	if err != nil {
			// 		resp = "Error for Merge: " + string(err.Error())
			// 		log.Printf("\t|... %s", resp)
			// 	} else {
			// 		// if enc && (conf.OwnerPW != "" || conf.UserPW != "") {
			// 		// 	err = api.EncryptFile(outfile, "", conf)
			// 		// 	if err != nil {
			// 		// 		resp = "Error for Encrypt: " + string(err.Error())
			// 		// 		log.Printf("\t|... %s", resp)
			// 		// 	}
			// 		// }
			// 		w, _ := os.Create(outfile)
			// 		// if err != nil {
			// 		// 	log.Fatal(err)
			// 		// }

			// 		defer func() {
			// 			// if err = w.Close(); err != nil {
			// 			// 	return
			// 			// }
			// 			w.Close()
			// 		}()
			// 		w.Write(reply.W)
			// 	}
			// }

		}
	}

	log.Printf("validate %s!\n", resp)
	fmt.Fprint(w, resp)
}

func main() {
	// http.HandleFunc("/", index)
	http.HandleFunc("/pdf/merge", pdfmerge)
	http.HandleFunc("/pdf/merge/", pdfmerge)
	http.HandleFunc("/pdf/validate", pdfvalidate)
	http.HandleFunc("/pdf/validate/", pdfvalidate)
	serv := "0.0.0.0:8384"
	if err := http.ListenAndServe(serv, nil); err != nil {
		log.Fatal("ListenAndServe: ", err)
	}
}
