package commands

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/daedaleanai/git-ticket/bug"
	"github.com/daedaleanai/git-ticket/cache"
	_select "github.com/daedaleanai/git-ticket/commands/select"
	"github.com/daedaleanai/git-ticket/util/colors"
	"github.com/daedaleanai/git-ticket/util/interrupt"
	"github.com/spf13/cobra"
)

var (
	showFieldsQuery string
)

func runShowBug(cmd *cobra.Command, args []string) error {
	backend, err := cache.NewRepoCache(repo)
	if err != nil {
		return err
	}
	defer backend.Close()
	interrupt.RegisterCleaner(backend.Close)

	b, args, err := _select.ResolveBug(backend, args)
	if err != nil {
		return err
	}

	snapshot := b.Snapshot()

	// process assignee
	assigneeName := "UNASSIGNED"
	if snapshot.Assignee != nil {
		assignee, err := backend.ResolveIdentityExcerpt(snapshot.Assignee.Id())
		if err != nil {
			return err
		}
		assigneeName = assignee.DisplayName()
	}

	// process labels
	var labels []string
	var workflow string = "<NONE ASSIGNED>"

	for _, lbl := range snapshot.Labels {
		if strings.HasPrefix(lbl.String(), "workflow:") {
			workflow = strings.TrimPrefix(lbl.String(), "workflow:")
		} else if strings.HasPrefix(lbl.String(), "checklist:") {
			continue
		} else {
			labels = append(labels, lbl.String())
		}
	}

	if len(snapshot.Comments) == 0 {
		return errors.New("invalid ticket: no comment")
	}

	firstComment := snapshot.Comments[0]

	if showFieldsQuery != "" {
		switch showFieldsQuery {
		case "assignee":
			fmt.Printf("%s\n", assigneeName)
		case "author":
			fmt.Printf("%s\n", firstComment.Author.DisplayName())
		case "authorEmail":
			fmt.Printf("%s\n", firstComment.Author.Email())
		case "createTime":
			fmt.Printf("%s\n", firstComment.FormatTime())
		case "humanId":
			fmt.Printf("%s\n", snapshot.Id().Human())
		case "id":
			fmt.Printf("%s\n", snapshot.Id())
		case "workflow":
			fmt.Printf("%s\n", workflow)
		case "checklists":
			for _, clMap := range snapshot.Checklists {
				for user, cl := range clMap {
					reviewer, err := backend.ResolveIdentityExcerpt(user)
					if err != nil {
						return err
					}
					fmt.Printf("%s reviewed %s: %s\n", reviewer.DisplayName(), cl.LastEdit, cl)
				}
			}
		case "reviews":
			for _, r := range snapshot.Reviews {
				// The Differential ID
				fmt.Printf("==== %s ====\n", r.RevisionId)

				// The statuses
				for u, s := range r.LatestUserStatuses() {
					var userName string

					if user, err := backend.ResolveIdentityPhabID(u); err != nil {
						userName = "???"
					} else {
						userName = user.DisplayName()
					}

					fmt.Printf("(%s) %-20s: %s\n", time.Unix(s.Timestamp, 0).Format(time.RFC822), userName, s.Status)
				}

				// Output all the comments
				fmt.Printf("---- %d comments ----\n", len(r.Comments))

				for _, c := range r.Comments {
					var userName string

					if user, err := backend.ResolveIdentityPhabID(c.User); err != nil {
						userName = "???"
					} else {
						userName = user.DisplayName()
					}

					fmt.Printf("(%s) %-20s: %s\n", time.Unix(c.Timestamp, 0).Format(time.RFC822), userName, c.OneLineComment())
				}
			}
		case "labels":
			for _, l := range labels {
				fmt.Printf("%s\n", l)
			}
		case "actors":
			for _, a := range snapshot.Actors {
				fmt.Printf("%s\n", a.DisplayName())
			}
		case "participants":
			for _, p := range snapshot.Participants {
				fmt.Printf("%s\n", p.DisplayName())
			}
		case "shortId":
			fmt.Printf("%s\n", snapshot.Id().Human())
		case "status":
			fmt.Printf("%s\n", snapshot.Status)
		case "title":
			fmt.Printf("%s\n", snapshot.Title)
		default:
			return fmt.Errorf("unsupported field: %s", showFieldsQuery)
		}

		return nil
	}

	// Header
	fmt.Printf("[%s] %s %s - %s\n\n",
		colors.Yellow(snapshot.Status),
		colors.Cyan(snapshot.Id().Human()),
		snapshot.Title,
		colors.Blue(assigneeName),
	)

	fmt.Printf("%s opened this issue %s\n\n",
		colors.Magenta(firstComment.Author.DisplayName()),
		firstComment.FormatTimeRel(),
	)

	// Workflow
	fmt.Printf("workflow: %s\n", workflow)

	// Checklists
	fmt.Printf("checklists:\n")
	for clLabel, st := range snapshot.GetChecklistCompoundStates() {
		cl, present := bug.ChecklistStore[clLabel]

		if !present {
			return fmt.Errorf("unknown checklist: %s", clLabel)
		}

		fmt.Printf("- %s : %s\n", cl.Title, st.ColorString())
	}

	// Reviews
	fmt.Printf("reviews:\n")
	for _, review := range snapshot.Reviews {
		fmt.Printf("- %s (%d comments)\n", review.RevisionId, len(review.Comments))
	}

	// Labels
	fmt.Printf("labels: %s\n",
		strings.Join(labels, ", "),
	)

	// Actors
	var actors = make([]string, len(snapshot.Actors))
	for i := range snapshot.Actors {
		actors[i] = snapshot.Actors[i].DisplayName()
	}

	fmt.Printf("actors: %s\n",
		strings.Join(actors, ", "),
	)

	// Participants
	var participants = make([]string, len(snapshot.Participants))
	for i := range snapshot.Participants {
		participants[i] = snapshot.Participants[i].DisplayName()
	}

	fmt.Printf("participants: %s\n\n",
		strings.Join(participants, ", "),
	)

	// Comments
	indent := "  "

	for i, comment := range snapshot.Comments {
		var message string
		fmt.Printf("%s#%d %s <%s>\n\n",
			indent,
			i,
			comment.Author.DisplayName(),
			comment.Author.Email(),
		)

		if comment.Message == "" {
			message = colors.GreyBold("No description provided.")
		} else {
			message = comment.Message
		}

		fmt.Printf("%s%s\n\n\n",
			indent,
			message,
		)
	}

	return nil
}

var showCmd = &cobra.Command{
	Use:     "show [<id>]",
	Short:   "Display the details of a ticket.",
	PreRunE: loadRepo,
	RunE:    runShowBug,
}

func init() {
	RootCmd.AddCommand(showCmd)
	showCmd.Flags().StringVarP(&showFieldsQuery, "field", "f", "",
		"Select field to display. Valid values are [assignee,author,authorEmail,checklists,createTime,humanId,id,labels,reviews,shortId,status,title,workflow,actors,participants]")
}
