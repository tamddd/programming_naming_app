package function

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"html/template"
	"net/http"
	"os"
	"strconv"
	"strings"

	"github.com/google/generative-ai-go/genai"
	"github.com/gorilla/mux"
	"google.golang.org/api/option"
)

// Option 選択肢の構造体
type Option struct {
	Value string `json:"value"`
	Text  string `json:"text"`
}

// Function 問題の関数のデータ構造
type Function struct {
	ID          int64    `json:"id"`
	Title       string   `json:"title"`
	Description string   `json:"description"`
	Code        string   `json:"code"`
	Options     []Option `json:"options"`
	Answer      string   `json:"answer"`
}

type Service struct {
	DB *sql.DB
}

func NewService(db *sql.DB) *Service {
	return &Service{DB: db}
}

func (s *Service) NewFunctionForm(w http.ResponseWriter, r *http.Request) {
	tmpl, err := template.ParseFiles("templates/admin/new.html")
	if err != nil {
		http.Error(w, "Internal Server Error: "+err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	tmpl.Execute(w, nil)
}

func (s *Service) CreateFunction(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		http.Error(w, "Bad Request", http.StatusBadRequest)
		return
	}

	title := r.FormValue("title")
	description := r.FormValue("description")
	code := r.FormValue("code")
	answer := r.FormValue("answer")

	var options []Option
	for i := 1; i <= 3; i++ {
		val := r.FormValue(fmt.Sprintf("option%d_value", i))
		text := r.FormValue(fmt.Sprintf("option%d_text", i))
		if val != "" {
			options = append(options, Option{Value: val, Text: text})
		}
	}

	optionsJSON, err := json.Marshal(options)
	if err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	query := `INSERT INTO functions (title, description, code, options, answer, user_id, is_public) VALUES (?, ?, ?, ?, ?, ?, ?)`
	_, err = s.DB.Exec(query, title, description, code, optionsJSON, answer, 1, true)
	if err != nil {
		http.Error(w, "Database Error: "+err.Error(), http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, "/functions", http.StatusSeeOther)
}

func (s *Service) GenerateWithAI(w http.ResponseWriter, r *http.Request) {
	ctx := context.Background()
	apiKey := os.Getenv("GEMINI_API_KEY")
	if apiKey == "" {
		// 🚨 環境変数エラーはサーバーサイドログに出力する
		fmt.Println("Error: GEMINI_API_KEY not set")
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	client, err := genai.NewClient(ctx, option.WithAPIKey(apiKey))
	if err != nil {
		fmt.Println("Error creating AI client:", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	defer client.Close()

	model := client.GenerativeModel("gemini-2.5-flash")

	prompt := `プログラミングの関数命名問題を1つ生成してください。

以下のJSON形式で返してください：
{
  "title": "問題のタイトル",
  "description": "問題の説明文",
  "code": "function ???(引数) { コード }",
  "options": [
    {"value": "選択肢1の値", "text": "選択肢1の表示名"},
    {"value": "選択肢2の値", "text": "選択肢2の表示名"},
    {"value": "選択肢3の値", "text": "選択肢3の表示名"}
  ],
  "answer": "正解の値"
}

注意:
- codeには必ず ??? を含めてください
- optionsは3つ作成してください
- answerはoptionsのいずれかのvalueと一致させてください
- 日本語で生成してください`

	// --- 1. AI Content Generation ---
	resp, err := model.GenerateContent(ctx, genai.Text(prompt))
	if err != nil {
		fmt.Println("Gemini API error during GenerateContent:", err)
		http.Error(w, "AI generation failed", http.StatusInternalServerError)
		return
	}

	// --- 2. Response Validation & Extraction ---
	if len(resp.Candidates) == 0 || len(resp.Candidates[0].Content.Parts) == 0 {
		fmt.Println("AI response was empty or contained no parts.")
		http.Error(w, "No response from AI", http.StatusInternalServerError)
		return
	}

	responseText, ok := resp.Candidates[0].Content.Parts[0].(genai.Text)
	if !ok {
		fmt.Println("AI response part is not genai.Text.")
		http.Error(w, "AI response part type is invalid", http.StatusInternalServerError)
		return
	}

	// --- 3. Clean up AI Output (Robustness) ---
	// AIが出力する可能性のあるMarkdownブロック（例: ```json...```）を削除する
	rawJSON := strings.TrimSpace(string(responseText))
	rawJSON = strings.TrimPrefix(rawJSON, "```json")
	rawJSON = strings.TrimSuffix(rawJSON, "```")
	rawJSON = strings.TrimSpace(rawJSON)

	// --- 4. JSON Unmarshal (Validation) ---
	var generatedFunction Function
	if err := json.Unmarshal([]byte(rawJSON), &generatedFunction); err != nil {
		// 不正なJSONはサーバーログに記録（デバッグ用）
		fmt.Printf("JSON Unmarshal error: %v. Raw AI response: %s\n", err, rawJSON)
		http.Error(w, "AI returned malformed JSON", http.StatusInternalServerError)
		return
	}

	// --- 5. Final Response ---
	w.Header().Set("Content-Type", "application/json; charset=utf-8")

	// 💡 Goの慣用的なjson.Encoderを使用して、検証済みの構造体をクライアントに書き出す
	if err := json.NewEncoder(w).Encode(generatedFunction); err != nil {
		fmt.Println("Error encoding response JSON:", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
	}
}

// problems はすべての問題を保持するマップ
var problems = map[int64]Function{
	1: {
		ID:          1,
		Title:       "Addition",
		Description: "2つの数値を足し算する関数です。???に適切な名前を選んでください。",
		Code:        "function ???(a, b) {\n  return a + b;\n}",
		Options: []Option{
			{Value: "add", Text: "add"},
			{Value: "func", Text: "func"},
			{Value: "doSomething", Text: "doSomething"},
		},
		Answer: "add",
	},
	2: {
		ID:          2,
		Title:       "User Name",
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

func (s *Service) GetFunctionById(w http.ResponseWriter, r *http.Request) {
	id := mux.Vars(r)["id"]
	function, err := s.getFunctionById(id)
	if err != nil {
		http.NotFound(w, r)
		return
	}

	tmpl, err := template.ParseFiles("templates/function.html")
	if err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	tmpl.Execute(w, function)
}

func (s *Service) getFunctionById(id string) (Function, error) {
	if os.Getenv("ENV") == "test" {
		idInt, err := strconv.ParseInt(id, 10, 64)
		if err != nil {
			return Function{}, fmt.Errorf("invalid id format: %s", id)
		}
		if problem, ok := problems[idInt]; ok {
			return problem, nil
		}
		return Function{}, fmt.Errorf("problem with id %s not found", id)
	} else {
		query := `SELECT id, title, description, code, options, answer FROM functions WHERE id = ?`
		row := s.DB.QueryRow(query, id)
		var problem Function
		var optionsJSON []byte
		err := row.Scan(&problem.ID, &problem.Title, &problem.Description, &problem.Code, &optionsJSON, &problem.Answer)
		if err != nil {
			return Function{}, fmt.Errorf("problem with id %s not found", id)
		}
		err = json.Unmarshal(optionsJSON, &problem.Options)
		if err != nil {
			return Function{}, fmt.Errorf("failed to parse options: %w", err)
		}
		return problem, nil
	}
}

func (s *Service) getAllFunctions() ([]Function, error) {
	var allProblems []Function
	if os.Getenv("ENV") == "test" {
		for _, problem := range problems {
			allProblems = append(allProblems, problem)
		}
		return allProblems, nil
	} else {
		query := `SELECT id, title, description, code, options, answer FROM functions`
		rows, err := s.DB.Query(query)
		if err != nil {
			fmt.Println("Query error:", err)
			return nil, err
		}
		defer rows.Close()

		for rows.Next() {
			var problem Function
			var optionsJSON []byte
			err := rows.Scan(&problem.ID, &problem.Title, &problem.Description, &problem.Code, &optionsJSON, &problem.Answer)
			if err != nil {
				fmt.Println("Scan error:", err)
				return nil, err
			}
			err = json.Unmarshal(optionsJSON, &problem.Options)
			if err != nil {
				fmt.Println("JSON unmarshal error:", err)
				return nil, err
			}
			allProblems = append(allProblems, problem)
		}
		return allProblems, nil
	}
}

// GetAllFunctions はすべての問題のリストを返す
func (s *Service) GetAllFunctions(w http.ResponseWriter, r *http.Request) {
	functions, err := s.getAllFunctions()
	if err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	tmpl, err := template.ParseFiles("templates/functions.html")
	if err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	tmpl.Execute(w, functions)
}
