package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"strings"

	"github.com/pdfcpu/pdfcpu/pkg/api"
	"github.com/pdfcpu/pdfcpu/pkg/pdfcpu"
	"github.com/tidwall/gjson"
)

func index(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()
	fmt.Println("Form: ", r.Form)
	fmt.Println("Path: ", r.URL.Path)
	fmt.Println(r.Form["a"])
	fmt.Println(r.Form["b"])
	for k, v := range r.Form {
		fmt.Println(k, "=>", v, strings.Join(v, "-"))
	}
	fmt.Fprint(w, "It works !")
}

func pdfmerge(w http.ResponseWriter, r *http.Request) {

	var resp string
	body, _ := ioutil.ReadAll(r.Body)
	//output file name
	outfile := gjson.GetBytes(body, "outfile").String()
	// fmt.Println(outfile)

	// input file names
	var infiles []string
	result := []byte(gjson.GetBytes(body, "infiles").Raw)
	// fmt.Println(string(result))
	json.Unmarshal(result, &infiles)

	// Create new Merged pdf
	err := api.MergeCreateFile(infiles, outfile, nil)
	if err != nil {
		resp = "Error: " + string(err.Error())
	} else {
		resp = "Success"
	}

	// permission set
	if gjson.GetBytes(body, "protect").Int() == 1 {
		// fmt.Println("need protect!")
		pwd := gjson.GetBytes(body, "password").String()
		conf := pdfcpu.NewAESConfiguration("", pwd, 256)
		conf.Permissions = 204
		err := api.EncryptFile(outfile, "", conf)
		if err != nil {
			resp += "Error encrypt: " + string(err.Error())
		}
		// api.SetPermissionsFile(outfile,"",conf)
	}

	// resp = "{\"success\":true,\"message\":\"success!\"}"
	fmt.Fprint(w, resp)
}

func main() {
	http.HandleFunc("/", index)
	http.HandleFunc("/pdf/merge", pdfmerge)
	http.HandleFunc("/pdf/merge/", pdfmerge)

	if err := http.ListenAndServe("0.0.0.0:8384", nil); err != nil {
		log.Fatal("ListenAndServe: ", err)
	}
}
