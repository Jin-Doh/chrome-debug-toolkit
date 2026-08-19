package report

import (
	"fmt"
	"io"
	"sort"

	"github.com/jin-doh/chrome-debug-toolkit/internal/netlog"
)

type eventCount struct {
	Name  string
	Count int
}

type errorCount struct {
	Code  int
	Count int
}

func PrintSummary(writer io.Writer, summary *netlog.Summary) {
	if summary == nil {
		fmt.Fprintln(writer, "NetLog summary unavailable")
		return
	}
	fmt.Fprintln(writer, "NetLog summary")
	fmt.Fprintln(writer)
	fmt.Fprintf(writer, "File:         %s\n", summary.Path)
	fmt.Fprintf(writer, "Total events: %d\n", summary.TotalEvents)
	if summary.UnknownEvents > 0 {
		fmt.Fprintf(writer, "Unknown:      %d\n", summary.UnknownEvents)
	}

	events := make([]eventCount, 0, len(summary.EventCounts))
	for name, count := range summary.EventCounts {
		events = append(events, eventCount{Name: name, Count: count})
	}
	sort.Slice(events, func(i, j int) bool {
		if events[i].Count != events[j].Count {
			return events[i].Count > events[j].Count
		}
		return events[i].Name < events[j].Name
	})
	if len(events) > 0 {
		fmt.Fprintln(writer, "\nEvents")
		for _, item := range events {
			fmt.Fprintf(writer, "  %-32s %d\n", item.Name, item.Count)
		}
	}

	errors := make([]errorCount, 0, len(summary.ErrorCounts))
	for code, count := range summary.ErrorCounts {
		errors = append(errors, errorCount{Code: code, Count: count})
	}
	sort.Slice(errors, func(i, j int) bool {
		if errors[i].Count != errors[j].Count {
			return errors[i].Count > errors[j].Count
		}
		return errors[i].Code < errors[j].Code
	})
	if len(errors) > 0 {
		fmt.Fprintln(writer, "\nErrors")
		for _, item := range errors {
			fmt.Fprintf(writer, "  %-8d %d\n", item.Code, item.Count)
		}
	}
}
