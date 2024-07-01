package main

/*
credits:
https://kilb.tech/golang-templates
https://francoposa.io/resources/golang/golang-templates-1
https://stackoverflow.com/questions/41176355/go-template-name
https://mdaverde.com/posts/dynamic-template-creation-go
https://golang-examples.tumblr.com/post/87553422434/template-and-associated-templates
https://www.youtube.com/watch?v=k5wJv4XO7a0
*/

import (
	"errors"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"os"
)

func main() {
	dir, err := os.Getwd()
	if err != nil {
		fmt.Println("Can't get working dir. ", err)
		log.Fatal(err)
	}
	fmt.Println("Working directory: ", dir)

	pages := []string{"home", "about"}						// <== keys in the cache map
	cache := make(map[string]*template.Template)	// cache map stores associated templates

	funcMap := template.FuncMap{
		"dict": func(values ...interface{}) (map[string]interface{}, error) {
			if len(values)%2 != 0 {
				return nil, errors.New("invalid dict call")
			}
			dict := make(map[string]interface{}, len(values)/2)
			for i := 0; i < len(values); i += 2 {
				key, ok := values[i].(string)
				if !ok {
					return nil, errors.New("dict keys must be strings")
				}
				dict[key] = values[i+1]
			}
			return dict, nil
		},
	}

	for _, page := range pages {
		// create a new named template instance and add the function
		goTpl := template.New("layout.tmpl").Funcs(funcMap)
		// now parse all other associated templates into the instance
		// we just created.
		// this process is name: "composing a template"
		goTpl, err = goTpl.ParseFiles(dir+"/templates/layout.tmpl",
			dir+"/templates/sidebar.tmpl",
			dir+"/templates/pages/"+page+".tmpl",
			dir+"/templates/headline.tmpl")
		if err != nil {
			log.Fatal("error parsing files: ", err)
		}
		// page names ("home", "about") acts as keys
		cache[page] = goTpl
	}

	// goT := cache["about"]
	// fmt.Println(goT.DefinedTemplates())
	// ; defined templates are: "layout.tmpl", "sidebar.tmpl", "about.tmpl", "headline", "headline.tmpl", "sidebar", "main"
	// defined blocks are:                                                   "headline",                  "sidebar", "main",
	// defined files are:       "layout.tmpl", "sidebar.tmpl", "about.tmpl",             "headline.tmpl"

	// goT := cache["home"]
	// fmt.Println(goT.DefinedTemplates())
	// ; defined templates are: "sidebar", "main", "layout.tmpl", "sidebar.tmpl", "home.tmpl", "headline", "headline.tmpl"
	// defined blocks are:      "sidebar", "main",                                             "headline",                 
	// defined files are:                          "layout.tmpl", "sidebar.tmpl", "home.tmpl",             "headline.tmpl"

	// So; 
	//	- a go template is essentially an element of map[string]*Template;
	//	- by calling cache["key"] we extract ALL associated/composed templates at ones;
	//	- we compose this association ourselfs, because a go template is dynamic;
	//	- composing a template is population the map[string]*Template structure;
	//	- last but not least: the name "go template" covers a complicated creature.

	// block structure:		used in:								defined in:
	// layout.tmpl
	// |_ main						layout.tmpl							home.tmpl, about.tmpl
	//    |_ headline			home.tmpl, about.tmpl		headline.tmpl
	// |_ sidebar					layout.tmpl							sidebar.tmpl

	// goT := cache["home"]
	// fmt.Println(goT.Tree.Root.String())
	/* output:
	<!DOCTYPE html>
<html lang="en">
<head>
  <meta charset="UTF-8">
  <meta name="viewport" content="width=device-width, initial-scale=1.0">
  <title>Go Template Example</title>
</head>
<body>
  <aside>
    {{template "sidebar" .}}
  </aside>
  <main>
    {{template "main" .}}
  </main>
</body>
</html>

<!--
As you can see the file contains a simple HTML document structure. 
...
-->
*/


	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		err := cache["home"].Execute(w, map[string]interface{}{"name": "Jan"})
		if err != nil {
			log.Fatal("can't execute template. ", err)
		}
	})

	http.HandleFunc("/about", func(w http.ResponseWriter, r *http.Request) {
		err := cache["about"].Execute(w, nil)
		if err != nil {
			log.Fatal("can't execute template. ", err)
		}
	})

	// create & start a web server that will render the template
	fmt.Println("Serving...")
	http.ListenAndServe(":8080", nil)
}

/*
Let's assume we want to implement a re-usable headline component.
Both of our pages, home and about should make use of this component.
But there's a catch:
The home page should show the headline in black, while the
about page should display it in blue.
Also, of course, the text of the headline should correspond to the current page.

The first part is a bit tricky:
We have to implement a template helper function which can be used to pass
variable data from template to template.
The helper function will have the name 'dict' and it accepts a map object
with key type string and any value type.
The dict function will first check if the number of passed arguments is
dividable by 2. If not, it will break.
Then, it will create a new map where each key is mapped to it's value.
This might seem a bit strange at the beginning, but it will simply passing
data in sub templates. You'll see...

Unfortunately it's not possible to pass helper functions to the ParseFiles method.
In theory, we could add these helper functions after parsing to the created template instance. 
But this would be too late. The functions have to be registered before parsing.
That's why we had to create a template instance first (template.New("layout.tmpl"))
(with the name of our base template file layout.tmpl), 
then add the helper function, then parse the template files.

*/
