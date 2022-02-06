package main

import (
	"html/template"
	"io/ioutil"
	"log"
	"net/http"
	"regexp"
)

type Page struct {
	Title string
	Body  []byte
}

// Page構造体に対してsaveメソッドを定義してる(ファイルの保存)
func (p *Page) save() error {
	filename := p.Title + ".txt"
	return ioutil.WriteFile(filename, p.Body, 0600)
}

// string型の引数を取り、Pageのポインタを返す関数
func loadPage(title string) (*Page, error) {
	filename := title + ".txt"
	body, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, err
	}
	// ポインタで返す
	return &Page{Title: "test", Body: body}, nil
}

// ファイルを先に読み込んでおく
var templates = template.Must(template.ParseFiles("edit.html", "view.html"))

// テンプレートを描画する関数
func renderTemplate(w http.ResponseWriter, tmpl string, p *Page) {
	// 描画するtemplateを第２引数に明示的に渡している
	err := templates.ExecuteTemplate(w, tmpl+".html", p)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

// view.htmlにview/以下のページ名を渡す
func viewHandler(w http.ResponseWriter, r *http.Request, title string) {
	p, err := loadPage(title)
	if err != nil {
		http.Redirect(w, r, "/edit/"+title, http.StatusFound)
	}
	renderTemplate(w, "view", p)
}

// edit.htmlにedit/以下のページ名を渡す
func editHandler(w http.ResponseWriter, r *http.Request, title string) {
	p, err := loadPage(title)
	// ページが存在しない場合は、新規ページを作成するためのStructを作成
	// エラーを返している
	if err != nil {
		p = &Page{Title: title}
	}
	renderTemplate(w, "edit", p)
}

// 保存ボタンを押した時に呼び出される関数
func saveHandler(w http.ResponseWriter, r *http.Request, title string) {
	body := r.FormValue("body")
	p := &Page{Title: title, Body: []byte(body)}
	// 毎回、ファイルを上書きしてる
	err := p.save()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	// エラーがなければ、viewHandlerにリダイレクト
	http.Redirect(w, r, "/view/"+title, http.StatusFound)
}

//// URLのtitle部分を抽出する処理を共通化
// Regexpオブジェクトを作成
var validPath = regexp.MustCompile("^/(edit|save|view)/([a-zA-Z0-9]+)$")

// **Handlerの引数に、http.ResponseWriterとhttp.Request、titleを渡している
// http.HandlerFuncは、通常の関数をHTTPハンドラーとして扱うためのアダプター
func makeHandler(fn func(http.ResponseWriter, *http.Request, string)) http.HandlerFunc {

	return func(w http.ResponseWriter, r *http.Request) {
		m := validPath.FindStringSubmatch(r.URL.Path)
		//=> [/edit/test3 edit test3]
		if m == nil {
			http.NotFound(w, r)
			return
		}
		// m[2]には、URLのtitle部分が入っている
		fn(w, r, m[2])
	}
}

// HandleFuncの定義
// func HandleFunc(pattern string, handler func(ResponseWriter, *Request)) {
// 	DefaultServeMux.HandleFunc(pattern, handler)
// }

func main() {
	// /view/~であれば、http.ListenAndServeに行く前にviewHandlerを呼び出す
	http.HandleFunc("/view/", makeHandler(viewHandler))
	http.HandleFunc("/edit/", makeHandler(editHandler))
	http.HandleFunc("/save/", makeHandler(saveHandler))
	// ポート8080でサーバーを起動
	log.Fatal(http.ListenAndServe(":8080", nil))
}
