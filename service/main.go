package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/huayuego/wordfilter/trie"
)

type router struct {
}

func (ro *router) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	switch r.URL.Path {
	case "/":
		apiHelper(w)
	case "/v1/query": // 查找敏感词
		queryWords(w, r)
	case "/v1/black_words": // 敏感词
		blackWords(w, r)
	case "/v1/white_prefix_words": // 白名单（前缀）
		whitePrefixWords(w, r)
	case "/v1/white_suffix_words": // 白名单（后缀）
		whiteSuffixWords(w, r)
	default:
		notFound(w)
	}
}

func notFound(w http.ResponseWriter) {
	w.WriteHeader(http.StatusNotFound)
}

func apiHelper(w http.ResponseWriter) {
	help := make(map[string]string)
	help["/v1/query?q={text} [GET,POST] "] = "查找敏感词"

	help["/v1/black_words [GET]"] = "查看敏感词"
	help["/v1/black_words [POST]"] = "添加敏感词"
	help["/v1/black_words [DELETE]"] = "删除敏感词"

	help["/v1/white_prefix_words [GET]"] = "查看白名单（前缀）词组"
	help["/v1/white_prefix_words [POST]"] = "添加白名单（前缀）词组"

	help["/v1/white_suffix_words [GET]"] = "查看白名单（后缀）词组"
	help["/v1/white_suffix_words [POST]"] = "添加白名单（后缀）词组"

	serveJSON(w, help)
}

func queryWords(w http.ResponseWriter, r *http.Request) {
	paramName := "q"

	type resp struct {
		Code     int      `json:"code"`
		Error    string   `json:"error,omitempty"`
		Keywords []string `json:"keywords,omitempty"`
		Text     string   `json:"text,omitempty"`
	}

	text := ""
	if r.Method == "GET" {
		params, err := url.ParseQuery(r.URL.RawQuery)
		if err == nil {
			if q, ok := params[paramName]; ok {
				text = q[0]
			}
		} else {
			fmt.Println(err)
		}

	} else if r.Method == "POST" {
		text = r.FormValue(paramName)
	}

	res := resp{
		Keywords: []string{},
	}

	if text != "" {
		res.Code = 1
		ok, keyword, newText := trie.BlackTrie().Query(strings.ToLower(text))
		if ok {
			res.Keywords = keyword
			res.Text = newText
		}
	} else {
		res.Code = 0
		res.Error = "参数" + paramName + "不能为空"
	}
	serveJSON(w, res)
}

func blackWords(w http.ResponseWriter, r *http.Request) {
	if r.Method == "GET" {
		showBlackWords(w, r)
	} else if r.Method == "POST" {
		addBlackWords(w, r)
	} else if r.Method == "DELETE" {
		deleteBlackWords(w, r)
	}
}

func addBlackWords(w http.ResponseWriter, r *http.Request) {
	resp := make(map[string]interface{})
	q := r.FormValue("q")

	if q == "" {
		resp["code"] = 0
		resp["error"] = "参数q不能为空"
	} else {
		i := 0
		words := strings.Split(q, ",")
		for _, s := range words {
			trie.BlackTrie().Add(strings.ToLower(strings.Trim(s, " ")))
			i++
		}

		resp["code"] = 1
		resp["mess"] = fmt.Sprintf("共添加了%d个敏感词", i)
	}

	serveJSON(w, resp)
}

func deleteBlackWords(w http.ResponseWriter, r *http.Request) {
	resp := make(map[string]interface{})

	q := r.FormValue("q")
	if q == "" {
		body, err := ioutil.ReadAll(r.Body)
		if err == nil {
			data := make(map[string]string)
			err = json.Unmarshal(body, &data)
			if err == nil {
				if qq, ok := data["q"]; ok {
					q = qq
				}
			}
		}
	}

	if q == "" {
		resp["code"] = 0
		resp["error"] = "参数q不能为空"
	} else {
		i := 0
		words := strings.Split(q, ",")
		for _, s := range words {
			trie.BlackTrie().Del(strings.ToLower(strings.Trim(s, " ")))
			i++
		}

		resp["code"] = 1
		resp["mess"] = fmt.Sprintf("共删除了%d个敏感词", i)
	}
	serveJSON(w, resp)
}

func showBlackWords(w http.ResponseWriter, r *http.Request) {
	words := trie.BlackTrie().ReadAll()
	str := strings.Join(words, "\n")
	w.Header().Set("Server", "goo")
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(200)
	w.Write([]byte(str))
}

func whitePrefixWords(w http.ResponseWriter, r *http.Request) {
	if r.Method == "GET" {
		words := trie.WhitePrefixTrie().ReadAll()
		str := strings.Join(words, "\n")
		w.Header().Set("Server", "goo")
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.WriteHeader(200)
		w.Write([]byte(str))

	} else if r.Method == "POST" {
		resp := make(map[string]interface{})
		q := r.FormValue("q")
		op := r.FormValue("type")
		if op == "init" {
			trie.ClearWhitePrefixTrie()
		}

		if q == "" {
			resp["code"] = 0
			resp["error"] = "参数q不能为空"
		} else {
			i := 0
			words := strings.Split(q, ",")
			for _, s := range words {
				trie.WhitePrefixTrie().Add(strings.ToLower(strings.Trim(s, " ")))
				i++
			}

			resp["code"] = 1
			resp["mess"] = fmt.Sprintf("共添加了%d个白名称前缀词", i)
		}

		serveJSON(w, resp)
	}
}

func whiteSuffixWords(w http.ResponseWriter, r *http.Request) {
	if r.Method == "GET" {
		words := trie.WhiteSuffixTrie().ReadAll()
		str := strings.Join(words, "\n")
		w.Header().Set("Server", "goo")
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.WriteHeader(200)
		w.Write([]byte(str))

	} else if r.Method == "POST" {
		resp := make(map[string]interface{})
		q := r.FormValue("q")
		op := r.FormValue("type")
		if op == "init" {
			trie.ClearWhiteSuffixTrie()
		}
		if q == "" {
			resp["code"] = 0
			resp["error"] = "参数q不能为空"
		} else {
			i := 0
			words := strings.Split(q, ",")
			for _, s := range words {
				trie.WhiteSuffixTrie().Add(strings.ToLower(strings.Trim(s, " ")))
				i++
			}

			resp["code"] = 1
			resp["mess"] = fmt.Sprintf("共添加了%d个白名称后缀词", i)
		}

		serveJSON(w, resp)
	}
}

func serveJSON(w http.ResponseWriter, data interface{}) {
	w.Header().Set("Server", "goo")
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(200)

	content, err := json.Marshal(data)
	if err == nil {
		w.Write(content)
	} else {
		w.Write([]byte(`{"code":0, "error":"解析JSON出错"}`))
	}
}

func main() {
	ipAddr := ":8080"
	if len(os.Args) > 1 {
		ipAddr = os.Args[1]
	}

	trie.InitAllTrie()

	t := time.Now().Local().Format("2006-01-02 15:04:05 -0700")
	fmt.Printf("%s Listen %s\n", t, ipAddr)
	http.ListenAndServe(ipAddr, &router{})
}
