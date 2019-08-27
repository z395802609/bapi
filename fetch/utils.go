package fetch

import (
	"github.com/JhuangLab/butils/log"
	"os"
)

func createIOStream(of *os.File, outfn string) *os.File {
	var err error
	if outfn == "" {
		of = os.Stdout
	} else {
		of, err = os.Create(outfn)
		if err != nil {
			log.Fatalf("error: %v", err)
		}
		of.Name()
	}
	return of
}

func setQueryFromEnd(from int, size int, total int) (int, int) {
	end := from + size
	if end == -1 || end > total {
		end = total
	}
	if from < 0 {
		from = 0
	} else if from > total {
		from = total
	}
	if end < from {
		end = from
	}
	return from, end
}
