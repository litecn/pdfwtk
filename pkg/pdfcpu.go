package pkg

import (
	"errors"
	"io"
	"os"
	"path/filepath"
	"strconv"
	"time"

	"github.com/pdfcpu/pdfcpu/pkg/api"
	"github.com/pdfcpu/pdfcpu/pkg/log"
	"github.com/pdfcpu/pdfcpu/pkg/pdfcpu"
	"github.com/pdfcpu/pdfcpu/pkg/pdfcpu/model"
	"github.com/pdfcpu/pdfcpu/pkg/pdfcpu/validate"
)

func Merge(destFile string, inFiles []string, w io.Writer, conf *model.Configuration) error {

	if w == nil {
		return errors.New("pdfcpu: Merge: Please provide w")
	}

	if conf == nil {
		conf = model.NewDefaultConfiguration()
	}
	conf.Cmd = model.MERGECREATE
	// conf.ValidationMode = model.ValidationRelaxed

	if destFile != "" {
		conf.Cmd = model.MERGEAPPEND
	}
	if destFile == "" {
		destFile = inFiles[0]
		inFiles = inFiles[1:]
	}

	f, err := os.Open(destFile)
	if err != nil {
		return err
	}
	defer f.Close()

	log.CLI.Println("merging into " + destFile)

	ctxDest, _, _, err := readAndValidate(f, conf, time.Now())
	if err != nil {
		return err
	}

	if conf.CreateBookmarks {
		if err := pdfcpu.EnsureOutlines(ctxDest, filepath.Base(destFile), conf.Cmd == model.MERGEAPPEND); err != nil {
			return err
		}
	}

	ctxDest.EnsureVersionForWriting()

	for _, fName := range inFiles {
		if err := func() error {
			f, err := os.Open(fName)
			if err != nil {
				return err
			}
			defer f.Close()

			log.CLI.Println(fName)
			if err = appendTo(f, filepath.Base(fName), ctxDest); err != nil {
				return err
			}

			return nil

		}(); err != nil {
			return err
		}
	}

	// if err := api.OptimizeContext(ctxDest); err != nil {
	// 	return err
	// }

	return api.WriteContext(ctxDest, w)
}

func MergeRaw(rsc []io.ReadSeeker, w io.Writer, conf *model.Configuration) error {

	if rsc == nil {
		return errors.New("pdfcpu: MergeRaw: missing rsc")
	}

	if w == nil {
		return errors.New("pdfcpu: MergeRaw: missing w")
	}

	if conf == nil {
		conf = model.NewDefaultConfiguration()
	}
	conf.Cmd = model.MERGECREATE
	// conf.ValidationMode = model.ValidationRelaxed
	conf.CreateBookmarks = false

	ctxDest, _, _, err := readAndValidate(rsc[0], conf, time.Now())
	if err != nil {
		return err
	}

	ctxDest.EnsureVersionForWriting()

	for i, f := range rsc[1:] {
		if err = appendTo(f, strconv.Itoa(i), ctxDest); err != nil {
			return err
		}
	}

	// if err = api.OptimizeContext(ctxDest); err != nil {
	// 	return err
	// }

	return api.WriteContext(ctxDest, w)
}

func MergeCreateFile(inFiles []string, outFile string, conf *model.Configuration) (err error) {

	f, err := os.Create(outFile)
	if err != nil {
		return err
	}

	defer func() {
		cerr := f.Close()
		if err == nil {
			err = cerr
		}
	}()

	log.CLI.Printf("writing %s...\n", outFile)
	return Merge("", inFiles, f, conf)
}

func readAndValidate(rs io.ReadSeeker, conf *model.Configuration, from1 time.Time) (ctx *model.Context, dur1, dur2 float64, err error) {
	if ctx, err = api.ReadContext(rs, conf); err != nil {
		return nil, 0, 0, err
	}

	dur1 = time.Since(from1).Seconds()

	if conf.ValidationMode == model.ValidationNone {
		// Bypass validation
		return ctx, 0, 0, nil
	}

	from2 := time.Now()

	if err = validate.XRefTable(ctx.XRefTable); err != nil {
		return nil, 0, 0, err
	}

	dur2 = time.Since(from2).Seconds()

	return ctx, dur1, dur2, nil
}

func appendTo(rs io.ReadSeeker, fName string, ctxDest *model.Context) error {
	ctxSource, _, _, err := readAndValidate(rs, ctxDest.Configuration, time.Now())
	if err != nil {
		return err
	}

	// Merge source context into dest context.
	return pdfcpu.MergeXRefTables(fName, ctxSource, ctxDest)
}
