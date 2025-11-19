package function

import (
	"fmt"
	"html/template"
	"net/http"

	"github.com/gorilla/mux"
)

// Option 選択肢の構造体
type Option struct {
	Value string `json:"value"`
	Text  string `json:"text"`
}

// Function 問題の関数のデータ構造
type Function struct {
	ID          string   `json:"id"`
	Description string   `json:"description"`
	Code        string   `json:"code"`
	Options     []Option `json:"options"`
	Answer      string   `json:"answer"`
}

// problems はすべての問題を保持するマップ
var problems = map[string]Function{
	"1": {
		ID:          "1",
		Description: "2つの数値を足し算する関数です。???に適切な名前を選んでください。",
		Code:        "function ???(a, b) {\n  return a + b;\n}",
		Options: []Option{
			{Value: "add", Text: "add"},
			{Value: "func", Text: "func"},
			{Value: "doSomething", Text: "doSomething"},
		},
		Answer: "add",
	},
	"2": {
		ID:          "2",
		Description: "ユーザー情報を受け取り、フルネームを返す関数です。???に適切な名前を選んでください。",
		Code:        "function ???(user) {\n  return user.firstName + ' ' + user.lastName;\n}",
		Options: []Option{
			{Value: "getUserName", Text: "getUserName"},
			{Value: "processUser", Text: "processUser"},
			{Value: "stringfy", Text: "stringfy"},
		},
		Answer: "getUserName",
	},
}

func GetFunctionById(w http.ResponseWriter, r *http.Request) {
	fmt.Println("GetFunctionById")
	id := mux.Vars(r)["id"]
	function, err := getFunctionById(id)
	if err != nil {
		http.NotFound(w, r)
		return
	}

	// テンプレートをパース
	tmpl, err := template.ParseFiles("templates/function.html")
	if err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	// テンプレートにデータを埋め込んでレスポンスを生成
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	tmpl.Execute(w, function)
}

func getFunctionById(id string) (Function, error) {
	if problem, ok := problems[id]; ok {
		return problem, nil
	}
	return Function{}, fmt.Errorf("problem with id %s not found", id)
}

// GetAllFunctions はすべての問題のリストを返す
func getAllFunctions() []Function {
	var allProblems []Function
	// マップの順序は保証されないので、IDでソートするなどが必要な場合があるが、今回はそのまま追加
	for _, problem := range problems {
		allProblems = append(allProblems, problem)
	}
	return allProblems
}

func GetAllFunctions(w http.ResponseWriter, r *http.Request) {
	functions := getAllFunctions()
	tmpl, err := template.ParseFiles("templates/functions.html")
	if err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	tmpl.Execute(w, functions)
}
