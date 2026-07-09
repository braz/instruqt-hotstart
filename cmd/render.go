package cmd

import (
	"encoding/json"
	"fmt"
	"io"
	"text/tabwriter"
	"time"

	"github.com/eoinbrazil/instruqt-hotstart/instruqt"
)

// renderPools writes pools as an indented JSON array or a table.
func renderPools(w io.Writer, pools []instruqt.HotStartPool, asJSON bool) error {
	if asJSON {
		return writeJSON(w, pools)
	}
	tw := tabwriter.NewWriter(w, 0, 0, 2, ' ', 0)
	fmt.Fprintln(tw, "ID\tNAME\tTYPE\tSIZE\tSTATUS\tSTARTS_AT\tENDS_AT")
	for _, p := range pools {
		fmt.Fprintf(tw, "%s\t%s\t%s\t%d\t%s\t%s\t%s\n",
			p.ID, p.Name, p.Type, p.Size, p.Status, fmtTime(p.StartsAt), fmtTime(p.EndsAt))
	}
	return tw.Flush()
}

// renderPool writes a single pool (JSON object or one-row table).
func renderPool(w io.Writer, p *instruqt.HotStartPool, asJSON bool) error {
	if asJSON {
		return writeJSON(w, p)
	}
	return renderPools(w, []instruqt.HotStartPool{*p}, false)
}

func writeJSON(w io.Writer, v any) error {
	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	return enc.Encode(v)
}

func fmtTime(t *time.Time) string {
	if t == nil {
		return "-"
	}
	return t.Format(time.RFC3339)
}
