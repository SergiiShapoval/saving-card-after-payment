package main

import (
	"net/http"

	"github.com/johnfercher/maroto/v2"
	"github.com/johnfercher/maroto/v2/pkg/components/text"
	"github.com/johnfercher/maroto/v2/pkg/config"
	"github.com/johnfercher/maroto/v2/pkg/consts/align"
	"github.com/johnfercher/maroto/v2/pkg/props"
)

func main() {
	http.HandleFunc("/download", downloadHandler)
	http.Handle("/", http.FileServer(http.Dir("static")))

	http.ListenAndServe(":8080", nil)
}

func downloadHandler(w http.ResponseWriter, r *http.Request) {
	cfg := config.NewBuilder().
		//WithPageNumber().
		WithLeftMargin(10).
		WithTopMargin(15).
		WithRightMargin(10).
		Build()

	mrt := maroto.New(cfg)
	mrt.AddRows(text.NewRow(12, "Invoice", props.Text{
		Top:   3,
		Align: align.Left,
		Size:  20,
	}))
	document, err := mrt.Generate()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	fileName := "sample.pdf"

	w.Header().Set("Content-Disposition", "attachment; filename="+fileName)
	w.Write(document.GetBytes())
}
