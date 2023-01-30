/*
Copyright © 2021 NAME HERE <EMAIL ADDRESS>

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/
package cmd

import (
	"net/http"
	"os"
	"path/filepath"
	"text/template"

	"embed"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

//go:embed templates/*
var templateFiles embed.FS

var pages = map[string]string{
	"/": "templates/index.html",
}

// installCmd represents the install command
var serveCmd = &cobra.Command{
	Use:    "serve",
	Short:  "Serve web frontend",
	Hidden: true,
	Long:   ``,
	Args:   cobra.ExactArgs(0), // https://github.com/spf13/cobra/blob/master/user_guide.md
	Run: func(cmd *cobra.Command, args []string) {
		http.HandleFunc("/", indexHandler)

		http.FileServer(http.FS(templateFiles))
		log.Println("server started...")
		err := http.ListenAndServe(":8080", nil)
		if err != nil {
			panic(err)
		}
		//log.Fatal(http.ListenAndServe(":8080", nil))

	},
}

func init() {
	rootCmd.AddCommand(serveCmd)

}

/*
/workspaces
/workspaces/$name
/workspaces/$name/blocks
/workspaces/$name/blocks/$name


*/

func indexHandler(w http.ResponseWriter, r *http.Request) {

	page, ok := pages[r.URL.Path]
	if !ok {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	t, err := template.ParseFS(templateFiles, page)
	if err != nil {
		log.Print(err.Error())
		http.Error(w, "Internal Server Error", 500)
		return
	}

	data := map[string]interface{}{
		"userAgent": r.UserAgent(),
	}

	err = t.Execute(w, &data)
	if err != nil {
		log.Fatal(err)
	}

	// w.Header().Set("Content-Type", "text/html")

	// data := map[string]interface{}{
	// 	"userAgent": r.UserAgent(),
	// }

	// err = t.ExecuteTemplate(w, page, data)
	// if err != nil {
	// 	log.Print(err.Error())
	// 	http.Error(w, "Internal Server Error", 500)
	// }

}

func ParseTemplateFS(dir string) (*template.Template, error) {
	var paths []string
	err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() {
			paths = append(paths, path)
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return template.ParseFiles(paths...)
}

func workspacesHandler(w http.ResponseWriter, r *http.Request) {
	t, err := template.ParseFiles("templates/workspaces.html")
	if err != nil {
		log.Print(err.Error())
		http.Error(w, "Internal Server Error", 500)
		return
	}
	t.Execute(w, workspace)
}
