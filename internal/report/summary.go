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

func PrintSummary(writer io.Writer, summary *netlog.Summary) error {
	if summary == nil {
		return writef(writer, "NetLog summary unavailable\n")
	}
	if err := writef(writer, "NetLog summary\n\n"); err != nil {
		return err
	}
	if err := writef(writer, "File:         %s\n", summary.Path); err != nil {
		return err
	}
	if err := writef(writer, "Total events: %d\n", summary.TotalEvents); err != nil {
		return err
	}
	if summary.UnknownEvents > 0 {
		if err := writef(writer, "Unknown:      %d\n", summary.UnknownEvents); err != nil {
			return err
		}
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
		if err := writef(writer, "\nEvents\n"); err != nil {
			return err
		}
		for _, item := range events {
			if err := writef(writer, "  %-32s %d\n", item.Name, item.Count); err != nil {
				return err
			}
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
		if err := writef(writer, "\nErrors\n"); err != nil {
			return err
		}
		for _, item := range errors {
			if err := writef(writer, "  %-8d %d\n", item.Code, item.Count); err != nil {
				return err
			}
		}
	}
	return nil
}

func writef(writer io.Writer, format string, args ...interface{}) error {
	_, err := fmt.Fprintf(writer, format, args...)
	return err
}
