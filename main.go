package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/pdfcpu/pdfcpu/pkg/api"
	"github.com/pdfcpu/pdfcpu/pkg/pdfcpu"
	"github.com/tidwall/gjson"
)

func index(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()
	fmt.Println("Form: ", r.Form)
	fmt.Println("Path: ", r.URL.Path)
	for k, v := range r.Form {
		fmt.Println(k, "=>", v, strings.Join(v, "-"))
	}
	fmt.Fprint(w, "It works !")
}

func pdfmerge(w http.ResponseWriter, r *http.Request) {

	resp := "Success"
	body, _ := ioutil.ReadAll(r.Body)
	//output file name
	outfile := gjson.GetBytes(body, "outfile").String()
	log.Println(outfile)
	// fmt.Println(outfile)

	// input file names
	var infiles []string
	result := []byte(gjson.GetBytes(body, "infiles").Raw)
	// fmt.Println(string(result))
	json.Unmarshal(result, &infiles)

	// configureation
	conf := pdfcpu.NewDefaultConfiguration()

	// encrypt check and permission set
	enc := gjson.GetBytes(body, "protect").Int() == 1
	if enc {
		// fmt.Println("need protect!")
		pwd := gjson.GetBytes(body, "password").String()
		conf.OwnerPW = pwd
		conf.UserPW = ""
		conf.EncryptUsingAES = true
		conf.EncryptKeyLength = 256
		conf.Permissions = 204
		if len(pwd) > 0 {
			log.Println("\tneed protect!")
		}
	}

	//validate
	for _, v := range infiles {
		// fmt.Println(v)
		err := api.ValidateFile(v, conf)
		if err != nil {
			fmt.Printf("%s validate error!\n", v)
			resp = "has error"
		}
	}
	// verr := api.ValidateFiles(infiles, conf)
	// if verr != nil {
	// 	fmt.Printf("validate error: %s", verr)
	// }

	// Create new Merged or/and Encrypt pdf
	err := api.MergeCreateFile(infiles, outfile, conf)
	if err != nil {
		resp = "Error for Merge: " + string(err.Error())
		log.Println(resp)
	} else {
		if enc && (conf.OwnerPW != "" || conf.UserPW != "") {
			err = api.EncryptFile(outfile, "", conf)
			if err != nil {
				resp = "Error for Encrypt: " + string(err.Error())
				log.Println(resp)
			}
		}
	}
	if err := os.Chmod(outfile, 0666); err != nil {
		log.Println(err)
	}
	log.Printf("%s %s!\n", outfile, resp)
	fmt.Fprint(w, resp)
}

func main() {
	http.HandleFunc("/", index)
	http.HandleFunc("/pdf/merge", pdfmerge)
	http.HandleFunc("/pdf/merge/", pdfmerge)

	serv := "0.0.0.0:8384"
	if err := http.ListenAndServe(serv, nil); err != nil {
		log.Fatal("ListenAndServe: ", err)
	}
}
