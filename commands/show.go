package commands

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/spf13/cobra"

	"github.com/daedaleanai/git-ticket/bug"
	_select "github.com/daedaleanai/git-ticket/commands/select"
	"github.com/daedaleanai/git-ticket/util/colors"
)

type showOptions struct {
	fields   string
	format   string
	timeline bool
}

func newShowCommand() *cobra.Command {
	env := newEnv()
	options := showOptions{}

	cmd := &cobra.Command{
		Use:      "show [ID]",
		Short:    "Display the details of a ticket.",
		PreRunE:  loadBackend(env),
		PostRunE: closeBackend(env),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runShow(env, options, args)
		},
	}

	flags := cmd.Flags()
	flags.SortFlags = false

	flags.BoolVarP(&options.timeline, "timeline", "t", false,
		"Output the timeline of the ticket")
	flags.StringVarP(&options.fields, "field", "", "",
		"Select field to display. Valid values are [author,authorEmail,createTime,lastEdit,humanId,id,labels,shortId,status,title,actors,participants]")
	flags.StringVarP(&options.format, "format", "f", "default",
		"Select the output formatting style. Valid values are [default,json,org-mode]")

	return cmd
}

func runShow(env *Env, opts showOptions, args []string) error {
	b, args, err := _select.ResolveBug(env.backend, args)
	if err != nil {
		return err
	}

	snap := b.Snapshot()

	if opts.timeline {
		printTimeline(env, snap)

		return nil
	}

	if len(snap.Comments) == 0 {
		return errors.New("invalid ticket: no comment")
	}

	workflow, labels := workflowAndLabels(snap)
	if opts.fields != "" {
		switch opts.fields {
		case "assignee":
			env.out.Printf("%s\n", snap.Assignee.DisplayName())
		case "author":
			env.out.Printf("%s\n", snap.Author.DisplayName())
		case "authorEmail":
			env.out.Printf("%s\n", snap.Author.Email())
		case "createTime":
			env.out.Printf("%s\n", snap.CreateTime.String())
		case "lastEdit":
			env.out.Printf("%s\n", snap.EditTime().String())
		case "humanId":
			env.out.Printf("%s\n", snap.Id().Human())
		case "id":
			env.out.Printf("%s\n", snap.Id())
		case "workflow":
			env.out.Printf("%s\n", workflow)
		case "checklists":
			for _, clMap := range snap.Checklists {
				for user, cl := range clMap {
					reviewer, err := env.backend.ResolveIdentityExcerpt(user)
					if err != nil {
						return err
					}
					env.out.Printf("%s reviewed %s: %s\n", reviewer.DisplayName(), cl.LastEdit, cl)
				}
			}
		case "reviews":
			for _, r := range snap.Reviews {
				// The Differential ID
				env.out.Printf("==== %s ====\n", r.RevisionId)

				// The statuses
				for u, s := range r.LatestUserStatuses() {
					var userName string

					if user, err := env.backend.ResolveIdentityPhabID(u); err != nil {
						userName = "???"
					} else {
						userName = user.DisplayName()
					}

					env.out.Printf("(%s) %-20s: %s\n", time.Unix(s.Timestamp, 0).Format(time.RFC822), userName, s.Status)
				}

				// Output all the comments
				env.out.Printf("---- %d comments ----\n", len(r.Comments))

				for _, c := range r.Comments {
					var userName string

					if user, err := env.backend.ResolveIdentityPhabID(c.User); err != nil {
						userName = "???"
					} else {
						userName = user.DisplayName()
					}

					env.out.Printf("(%s) %-20s: %s\n", time.Unix(c.Timestamp, 0).Format(time.RFC822), userName, c.OneLineComment())
				}
			}
		case "labels":
			for _, l := range labels {
				env.out.Printf("%s\n", l)
			}
		case "actors":
			for _, a := range snap.Actors {
				env.out.Printf("%s\n", a.DisplayName())
			}
		case "participants":
			for _, p := range snap.Participants {
				env.out.Printf("%s\n", p.DisplayName())
			}
		case "shortId":
			env.out.Printf("%s\n", snap.Id().Human())
		case "status":
			env.out.Printf("%s\n", snap.Status)
		case "title":
			env.out.Printf("%s\n", snap.Title)
		default:
			return fmt.Errorf("\nUnsupported field: %s\n", opts.fields)
		}

		return nil
	}

	switch opts.format {
	case "org-mode":
		return showOrgModeFormatter(env, snap)
	case "json":
		return showJsonFormatter(env, snap)
	case "default":
		return showDefaultFormatter(env, snap)
	default:
		return fmt.Errorf("unknown format %s", opts.format)
	}
}

func workflowAndLabels(snap *bug.Snapshot) (string, []string) {
	var labels []string
	var workflow = "<NONE ASSIGNED>"

	for _, lbl := range snap.Labels {
		if lbl.IsWorkflow() {
			workflow = strings.TrimPrefix(lbl.String(), "workflow:")
		} else if lbl.IsChecklist() {
			continue
		} else {
			labels = append(labels, lbl.String())
		}
	}
	return workflow, labels
}

func showDefaultFormatter(env *Env, snapshot *bug.Snapshot) error {
	// Header
	env.out.Printf("%s [%s] %s - %s\n\n",
		colors.Cyan(snapshot.Id().Human()),
		colors.Yellow(snapshot.Status),
		snapshot.Title,
		colors.Blue(snapshot.Assignee.DisplayName()),
	)

	env.out.Printf("%s opened this issue %s\n",
		colors.Magenta(snapshot.Author.DisplayName()),
		snapshot.CreateTime.String(),
	)

	env.out.Printf("This was last edited at %s\n\n",
		snapshot.EditTime().String(),
	)

	// Workflow
	workflow, labels := workflowAndLabels(snapshot)
	env.out.Printf("workflow: %s\n", workflow)

	// Checklists
	env.out.Printf("checklists:\n")
	for clLabel, st := range snapshot.GetChecklistCompoundStates() {
		cl, err := bug.GetChecklist(clLabel)

		if err != nil {
			return err
		}

		env.out.Printf("- %s : %s\n", cl.Title, st.ColorString())
	}

	// Reviews
	env.out.Printf("reviews:\n")
	for _, review := range snapshot.Reviews {
		env.out.Printf("- %s (%d comments)\n", review.RevisionId, len(review.Comments))
	}

	// Labels
	env.out.Printf("labels: %s\n",
		strings.Join(labels, ", "),
	)

	// Actors
	var actors = make([]string, len(snapshot.Actors))
	for i := range snapshot.Actors {
		actors[i] = snapshot.Actors[i].DisplayName()
	}

	env.out.Printf("actors: %s\n",
		strings.Join(actors, ", "),
	)

	// Participants
	var participants = make([]string, len(snapshot.Participants))
	for i := range snapshot.Participants {
		participants[i] = snapshot.Participants[i].DisplayName()
	}

	env.out.Printf("participants: %s\n\n",
		strings.Join(participants, ", "),
	)

	// Comments
	indent := "  "

	for i, comment := range snapshot.Comments {
		var message string
		env.out.Printf("%s#%d %s <%s>\n\n",
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

		env.out.Printf("%s%s\n\n\n",
			indent,
			message,
		)
	}

	return nil
}

// printTimeline walks the snapshot timeline for the selected bug and prints
// some information
func printTimeline(env *Env, snapshot *bug.Snapshot) {
	for _, op := range snapshot.Timeline {
		switch op := op.(type) {
		case *bug.AddCommentTimelineItem:
			env.out.Printf("(%s) %-20s: commented \"%s\"\n", op.CreatedAt.Time().Format(time.RFC822), op.Author.DisplayName(), op.Message)

		case *bug.CreateTimelineItem:
			env.out.Printf("(%s) %-20s: created \"%s\"\n", op.CreatedAt.Time().Format(time.RFC822), op.Author.DisplayName(), snapshot.Title)

		case *bug.LabelChangeTimelineItem:
			if len(op.Added) > 0 {
				env.out.Printf("(%s) %-20s: added labels ", op.UnixTime.Time().Format(time.RFC822), op.Author.DisplayName())
				for _, label := range op.Added {
					env.out.Printf("\"%s\" ", label)
				}
				fmt.Println()
			}

			if len(op.Removed) > 0 {
				env.out.Printf("(%s) %-20s: removed labels ", op.UnixTime.Time().Format(time.RFC822), op.Author.DisplayName())
				for _, label := range op.Removed {
					env.out.Printf("\"%s\" ", label)
				}
				fmt.Println()
			}

		case *bug.SetAssigneeTimelineItem:
			env.out.Printf("(%s) %-20s: set assignee \"%s\"\n", op.UnixTime.Time().Format(time.RFC822), op.Author.DisplayName(), op.Assignee.DisplayName())

		case *bug.SetChecklistTimelineItem:
			env.out.Printf("(%s) %-20s: edited \"%s\"\n", op.UnixTime.Time().Format(time.RFC822), op.Author.DisplayName(), op.Checklist.Title)

		case *bug.SetStatusTimelineItem:
			env.out.Printf("(%s) %-20s: \"%s\"\n", op.UnixTime.Time().Format(time.RFC822), op.Author.DisplayName(), op.Status.Action())
		}

	}
}

type JSONBugSnapshot struct {
	Id           string         `json:"id"`
	HumanId      string         `json:"human_id"`
	CreateTime   JSONTime       `json:"create_time"`
	EditTime     JSONTime       `json:"edit_time"`
	Status       string         `json:"status"`
	Labels       []bug.Label    `json:"labels"`
	Title        string         `json:"title"`
	Author       JSONIdentity   `json:"author"`
	Actors       []JSONIdentity `json:"actors"`
	Participants []JSONIdentity `json:"participants"`
	Comments     []JSONComment  `json:"comments"`
}

type JSONComment struct {
	Id      string       `json:"id"`
	HumanId string       `json:"human_id"`
	Author  JSONIdentity `json:"author"`
	Message string       `json:"message"`
}

func NewJSONComment(comment bug.Comment) JSONComment {
	return JSONComment{
		Id:      comment.Id().String(),
		HumanId: comment.Id().Human(),
		Author:  NewJSONIdentity(comment.Author),
		Message: comment.Message,
	}
}

func showJsonFormatter(env *Env, snapshot *bug.Snapshot) error {
	jsonBug := JSONBugSnapshot{
		Id:         snapshot.Id().String(),
		HumanId:    snapshot.Id().Human(),
		CreateTime: NewJSONTime(snapshot.CreateTime, 0),
		EditTime:   NewJSONTime(snapshot.EditTime(), 0),
		Status:     snapshot.Status.String(),
		Labels:     snapshot.Labels,
		Title:      snapshot.Title,
		Author:     NewJSONIdentity(snapshot.Author),
	}

	jsonBug.Actors = make([]JSONIdentity, len(snapshot.Actors))
	for i, element := range snapshot.Actors {
		jsonBug.Actors[i] = NewJSONIdentity(element)
	}

	jsonBug.Participants = make([]JSONIdentity, len(snapshot.Participants))
	for i, element := range snapshot.Participants {
		jsonBug.Participants[i] = NewJSONIdentity(element)
	}

	jsonBug.Comments = make([]JSONComment, len(snapshot.Comments))
	for i, comment := range snapshot.Comments {
		jsonBug.Comments[i] = NewJSONComment(comment)
	}

	jsonObject, _ := json.MarshalIndent(jsonBug, "", "    ")
	env.out.Printf("%s\n", jsonObject)

	return nil
}

func showOrgModeFormatter(env *Env, snapshot *bug.Snapshot) error {
	// Header
	env.out.Printf("%s [%s] %s\n",
		snapshot.Id().Human(),
		snapshot.Status,
		snapshot.Title,
	)

	env.out.Printf("* Author: %s\n",
		snapshot.Author.DisplayName(),
	)

	env.out.Printf("* Creation Time: %s\n",
		snapshot.CreateTime.String(),
	)

	env.out.Printf("* Last Edit: %s\n",
		snapshot.EditTime().String(),
	)

	// Labels
	var labels = make([]string, len(snapshot.Labels))
	for i, label := range snapshot.Labels {
		labels[i] = string(label)
	}

	env.out.Printf("* Labels:\n")
	if len(labels) > 0 {
		env.out.Printf("** %s\n",
			strings.Join(labels, "\n** "),
		)
	}

	// Actors
	var actors = make([]string, len(snapshot.Actors))
	for i, actor := range snapshot.Actors {
		actors[i] = fmt.Sprintf("%s %s",
			actor.Id().Human(),
			actor.DisplayName(),
		)
	}

	env.out.Printf("* Actors:\n** %s\n",
		strings.Join(actors, "\n** "),
	)

	// Participants
	var participants = make([]string, len(snapshot.Participants))
	for i, participant := range snapshot.Participants {
		participants[i] = fmt.Sprintf("%s %s",
			participant.Id().Human(),
			participant.DisplayName(),
		)
	}

	env.out.Printf("* Participants:\n** %s\n",
		strings.Join(participants, "\n** "),
	)

	env.out.Printf("* Comments:\n")

	for i, comment := range snapshot.Comments {
		var message string
		env.out.Printf("** #%d %s\n",
			i, comment.Author.DisplayName())

		if comment.Message == "" {
			message = "No description provided."
		} else {
			message = strings.ReplaceAll(comment.Message, "\n", "\n: ")
		}

		env.out.Printf(": %s\n", message)
	}

	return nil
}
