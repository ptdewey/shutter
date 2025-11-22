package review

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/ptdewey/shutter/internal/diff"
	"github.com/ptdewey/shutter/internal/files"
	"github.com/ptdewey/shutter/internal/pretty"
)

type ReviewChoice int

const (
	Accept ReviewChoice = iota
	Reject
	Skip
	AcceptAllChoice
	RejectAllChoice
	SkipAllChoice
	Quit
)

func computeDiffLines(old, new *files.Snapshot) []diff.DiffLine {
	return diff.Histogram(old.Content, new.Content)
}

func Review() error {
	snapshots, err := files.ListNewSnapshots()
	if err != nil {
		return err
	}

	if len(snapshots) == 0 {
		fmt.Println(pretty.Success("✓ No new snapshots to review"))
		return nil
	}

	fmt.Println(pretty.Header("Review Snapshots"))
	fmt.Printf("Found %d new snapshot(s) to review\n\n", len(snapshots))

	return reviewLoop(snapshots)
}

func reviewLoop(snapshots []string) error {
	reader := bufio.NewReader(os.Stdin)

	for i, testName := range snapshots {
		fmt.Printf("\n[%d/%d] %s\n", i+1, len(snapshots), pretty.Header(testName))

		newSnap, err := files.ReadSnapshot(testName, "new")
		if err != nil {
			fmt.Println(pretty.Error("✗ Failed to read new snapshot: " + err.Error()))
			continue
		}

		accepted, acceptErr := files.ReadSnapshot(testName, "accepted")

		if acceptErr == nil {
			diffLines := computeDiffLines(accepted, newSnap)
			fmt.Println(pretty.DiffSnapshotBox(accepted, newSnap, diffLines))
		} else {
			fmt.Println(pretty.NewSnapshotBox(newSnap))
		}

		for {
			choice, err := askChoice(reader, i+1, len(snapshots))
			if err != nil {
				return err
			}

			switch choice {
			case Accept:
				if err := files.AcceptSnapshot(testName); err != nil {
					fmt.Println(pretty.Error("✗ Failed to accept snapshot: " + err.Error()))
				} else {
					fmt.Println(pretty.Success("✓ Snapshot accepted"))
				}
			case Reject:
				if err := files.RejectSnapshot(testName); err != nil {
					fmt.Println(pretty.Error("✗ Failed to reject snapshot: " + err.Error()))
				} else {
					fmt.Println(pretty.Warning("⊘ Snapshot rejected"))
				}
			case Skip:
				fmt.Println(pretty.Warning("⊘ Snapshot skipped"))
			case AcceptAllChoice:
				for j := i; j < len(snapshots); j++ {
					if err := files.AcceptSnapshot(snapshots[j]); err != nil {
						fmt.Println(pretty.Error("✗ Failed to accept snapshot: " + err.Error()))
					}
				}
				fmt.Printf(pretty.Success("✓ Accepted %d snapshot(s)\n"), len(snapshots)-i)
				return nil
			case RejectAllChoice:
				for j := i; j < len(snapshots); j++ {
					if err := files.RejectSnapshot(snapshots[j]); err != nil {
						fmt.Println(pretty.Error("✗ Failed to reject snapshot: " + err.Error()))
					}
				}
				fmt.Printf(pretty.Warning("⊘ Rejected %d snapshot(s)\n"), len(snapshots)-i)
				return nil
			case SkipAllChoice:
				fmt.Printf(pretty.Warning("⊘ Skipped %d snapshot(s)\n"), len(snapshots)-i)
				return nil
			case Quit:
				fmt.Println("\nReview interrupted")
				return nil
			}
			break
		}
	}

	fmt.Println("\n" + pretty.Success("✓ Review complete"))
	return nil
}

func askChoice(reader *bufio.Reader, current, total int) (ReviewChoice, error) {
	fmt.Printf("\nOptions: [a]ccept [r]eject [s]kip [A]ccept All [R]eject All [S]kip All [q]uit: ")

	input, err := reader.ReadString('\n')
	if err != nil {
		return Quit, err
	}

	input = strings.TrimSpace(input)

	switch input {
	case "a", "accept":
		return Accept, nil
	case "r", "reject":
		return Reject, nil
	case "s", "skip":
		return Skip, nil
	case "A", "Accept All":
		return AcceptAllChoice, nil
	case "R", "Reject All":
		return RejectAllChoice, nil
	case "S", "Skip All":
		return SkipAllChoice, nil
	case "q", "quit":
		return Quit, nil
	default:
		fmt.Println(pretty.Warning("Invalid option, please try again"))
		return askChoice(reader, current, total)
	}
}

func AcceptAll() error {
	snapshots, err := files.ListNewSnapshots()
	if err != nil {
		return err
	}

	for _, testName := range snapshots {
		if err := files.AcceptSnapshot(testName); err != nil {
			return err
		}
	}

	fmt.Printf(pretty.Success("✓ Accepted %d snapshot(s)\n"), len(snapshots))
	return nil
}

func RejectAll() error {
	snapshots, err := files.ListNewSnapshots()
	if err != nil {
		return err
	}

	for _, testName := range snapshots {
		if err := files.RejectSnapshot(testName); err != nil {
			return err
		}
	}

	fmt.Printf(pretty.Warning("⊘ Rejected %d snapshot(s)\n"), len(snapshots))
	return nil
}
