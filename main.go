package main

import (
	"fmt"
	"html/template"
	"log"
	"net/http"
	"strconv"
	"strings"
)

type PageData struct {
	Title   string
	Mode    string // "1" or "2"
	Values  map[string]string
	Result  *ResultData
	Error   string
	Version string
}

type ResultData struct {
	// Task 1
	KRS   float64
	KRG   float64
	HS    float64
	CS    float64
	SS    float64
	NS    float64
	OS    float64
	AS    float64
	HG    float64
	CG    float64
	SG    float64
	NG    float64
	OG    float64
	QrMJ  float64
	QsMJ  float64
	QgMJ  float64

	// Task 2
	KGR   float64
	HR    float64
	CR    float64
	SR    float64
	OR    float64
	AR    float64
	VRmg  float64
	Qr2MJ float64
}

var (
	tpl *template.Template
)

func main() {
	var err error
	tpl, err = template.ParseGlob("templates/*.html")
	if err != nil {
		log.Fatalf("parse templates: %v", err)
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/", indexHandler)
	mux.HandleFunc("/calculate", calculateHandler)
	mux.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("static"))))

	addr := ":8080"
	log.Printf("Listening on http://localhost%s ...", addr)
	log.Fatal(http.ListenAndServe(addr, withNoCache(mux)))
}

func withNoCache(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Cache-Control", "no-store")
		next.ServeHTTP(w, r)
	})
}

func indexHandler(w http.ResponseWriter, r *http.Request) {
	mode := r.URL.Query().Get("mode")
	if mode != "2" {
		mode = "1"
	}
	data := PageData{
		Title:   "Калькулятор палива (Go Web)",
		Mode:    mode,
		Values:  map[string]string{},
		Result:  nil,
		Error:   "",
		Version: "v1.0",
	}
	render(w, "index.html", data)
}

func calculateHandler(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}
	mode := r.FormValue("mode")
	if mode != "2" {
		mode = "1"
	}

	values := map[string]string{}
	for _, k := range []string{"H", "C", "S", "N", "O", "W", "A", "V", "Qg"} {
		values[k] = r.FormValue(k)
	}

	data := PageData{
		Title:   "Калькулятор палива (Go Web)",
		Mode:    mode,
		Values:  values,
		Result:  nil,
		Error:   "",
		Version: "v1.0",
	}

	if mode == "1" {
		res, errMsg := calcTask1(values)
		if errMsg != "" {
			data.Error = errMsg
		} else {
			data.Result = res
		}
	} else {
		res, errMsg := calcTask2(values)
		if errMsg != "" {
			data.Error = errMsg
		} else {
			data.Result = res
		}
	}

	render(w, "index.html", data)
}

func render(w http.ResponseWriter, name string, data PageData) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if err := tpl.ExecuteTemplate(w, name, data); err != nil {
		log.Printf("template exec error: %v", err)
		http.Error(w, "template error", http.StatusInternalServerError)
	}
}

func parsePercent(values map[string]string, key string, required bool) (float64, string) {
	raw := strings.TrimSpace(values[key])
	if raw == "" {
		if required {
			return 0, fmt.Sprintf("Поле %s є обов'язковим.", key)
		}
		return 0, ""
	}
	raw = strings.ReplaceAll(raw, ",", ".")
	v, err := strconv.ParseFloat(raw, 64)
	if err != nil {
		return 0, fmt.Sprintf("Поле %s має бути числом.", key)
	}
	return v, ""
}

func calcTask1(values map[string]string) (*ResultData, string) {
	H, e := parsePercent(values, "H", true); if e != "" { return nil, e }
	C, e := parsePercent(values, "C", true); if e != "" { return nil, e }
	S, e := parsePercent(values, "S", true); if e != "" { return nil, e }
	N, e := parsePercent(values, "N", true); if e != "" { return nil, e }
	O, e := parsePercent(values, "O", true); if e != "" { return nil, e }
	W, e := parsePercent(values, "W", true); if e != "" { return nil, e }
	A, e := parsePercent(values, "A", true); if e != "" { return nil, e }

	if W >= 100 {
		return nil, "W має бути < 100."
	}
	if W+A >= 100 {
		return nil, "W + A має бути < 100."
	}

	krs := 100.0 / (100.0 - W)
	krg := 100.0 / (100.0 - W - A)

	// dry mass
	hS := H * krs
	cS := C * krs
	sS := S * krs
	nS := N * krs
	oS := O * krs
	aS := A * krs

	// combustible mass
	hG := H * krg
	cG := C * krg
	sG := S * krg
	nG := N * krg
	oG := O * krg

	// lower heating value (kJ/kg)
	qr_kJ := 339.0*C + 1030.0*H - 108.8*(O-S) - 25.0*W
	qs_kJ := (qr_kJ + 0.025*W) * 100.0 / (100.0 - W)
	qg_kJ := (qr_kJ + 0.025*W) * 100.0 / (100.0 - W - A)

	return &ResultData{
		KRS:  krs,
		KRG:  krg,
		HS:   hS,
		CS:   cS,
		SS:   sS,
		NS:   nS,
		OS:   oS,
		AS:   aS,
		HG:   hG,
		CG:   cG,
		SG:   sG,
		NG:   nG,
		OG:   oG,
		QrMJ: qr_kJ / 1000.0,
		QsMJ: qs_kJ / 1000.0,
		QgMJ: qg_kJ / 1000.0,
	}, ""
}

func calcTask2(values map[string]string) (*ResultData, string) {
	H, e := parsePercent(values, "H", true); if e != "" { return nil, e }
	C, e := parsePercent(values, "C", true); if e != "" { return nil, e }
	S, e := parsePercent(values, "S", true); if e != "" { return nil, e }
	O, e := parsePercent(values, "O", true); if e != "" { return nil, e }
	W, e := parsePercent(values, "W", true); if e != "" { return nil, e }
	A, e := parsePercent(values, "A", true); if e != "" { return nil, e }
	V, e := parsePercent(values, "V", true); if e != "" { return nil, e } // mg/kg
	Qg, e := parsePercent(values, "Qg", true); if e != "" { return nil, e } // MJ/kg

	if W+A >= 100 {
		return nil, "W + A має бути < 100."
	}

	kgr := (100.0 - W - A) / 100.0

	hR := H * kgr
	cR := C * kgr
	sR := S * kgr
	oR := O * kgr
	aR := A * kgr
	vR := V * kgr

	qr := Qg*(100.0-W-aR)/100.0 - 0.025*W

	return &ResultData{
		KGR:   kgr,
		HR:    hR,
		CR:    cR,
		SR:    sR,
		OR:    oR,
		AR:    aR,
		VRmg:  vR,
		Qr2MJ: qr,
	}, ""
}
