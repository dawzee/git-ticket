package commands

import (
	"fmt"
	"strings"

	text "github.com/MichaelMure/go-term-text"
	"github.com/spf13/cobra"

	"github.com/daedaleanai/git-ticket/cache"
	"github.com/daedaleanai/git-ticket/util/colors"
	"github.com/daedaleanai/git-ticket/util/interrupt"
)

var (
	lsStatusQuery      []string
	lsAuthorQuery      []string
	lsAssigneeQuery    []string
	lsParticipantQuery []string
	lsLabelQuery       []string
	lsTitleQuery       []string
	lsActorQuery       []string
	lsNoQuery          []string
	lsSortBy           string
	lsSortDirection    string
)

func runLsBug(cmd *cobra.Command, args []string) error {
	backend, err := cache.NewRepoCache(repo)
	if err != nil {
		return err
	}
	defer backend.Close()
	interrupt.RegisterCleaner(backend.Close)

	var query *cache.Query
	if len(args) >= 1 {
		query, err = cache.ParseQuery(strings.Join(args, " "))

		if err != nil {
			return err
		}
	} else {
		query, err = lsQueryFromFlags()
		if err != nil {
			return err
		}
	}

	allIds := backend.QueryBugs(query)

	for _, id := range allIds {
		b, err := backend.ResolveBugExcerpt(id)
		if err != nil {
			return err
		}

		var authorName string
		if b.AuthorId != "" {
			author, err := backend.ResolveIdentityExcerpt(b.AuthorId)
			if err != nil {
				authorName = "<missing author data>"
			} else {
				authorName = author.DisplayName()
			}
		} else {
			authorName = b.LegacyAuthor.DisplayName()
		}

		assigneeName := "UNASSIGNED"
		if b.AssigneeId != "" {
			assignee, err := backend.ResolveIdentityExcerpt(b.AssigneeId)
			if err != nil {
				return err
			}
			assigneeName = assignee.DisplayName()
		}

		var labelsTxt strings.Builder
		for _, l := range b.Labels {
			lc256 := l.Color().Term256()
			labelsTxt.WriteString(lc256.Escape())
			labelsTxt.WriteString(" ◼")
			labelsTxt.WriteString(lc256.Unescape())
		}

		// truncate + pad if needed
		labelsFmt := text.TruncateMax(labelsTxt.String(), 10)
		titleFmt := text.LeftPadMaxLine(b.Title, 50-text.Len(labelsFmt), 0)
		authorFmt := text.LeftPadMaxLine(authorName, 15, 0)
		assigneeFmt := text.LeftPadMaxLine(assigneeName, 15, 0)

		comments := fmt.Sprintf("%4d 💬", b.LenComments)
		if b.LenComments > 9999 {
			comments = "    ∞ 💬"
		}

		fmt.Printf("%s %s\t%s\t%s\t%s\t%s\n",
			colors.Cyan(b.Id.Human()),
			text.LeftPadMaxLine(colors.Yellow(b.Status), 10, 0),
			titleFmt+labelsFmt,
			colors.Magenta(authorFmt),
			colors.Blue(assigneeFmt),
			comments,
		)
	}

	return nil
}

// Transform the command flags into a query
func lsQueryFromFlags() (*cache.Query, error) {
	query := cache.NewQuery()

	for _, status := range lsStatusQuery {
		f, err := cache.StatusFilter(status)
		if err != nil {
			return nil, err
		}
		query.Status = append(query.Status, f)
	}

	for _, title := range lsTitleQuery {
		f := cache.TitleFilter(title)
		query.Title = append(query.Title, f)
	}

	for _, author := range lsAuthorQuery {
		f := cache.AuthorFilter(author)
		query.Author = append(query.Author, f)
	}

	for _, assignee := range lsAssigneeQuery {
		f := cache.AssigneeFilter(assignee)
		query.Assignee = append(query.Assignee, f)
	}

	for _, actor := range lsActorQuery {
		f := cache.ActorFilter(actor)
		query.Actor = append(query.Actor, f)
	}

	for _, participant := range lsParticipantQuery {
		f := cache.ParticipantFilter(participant)
		query.Participant = append(query.Participant, f)
	}

	for _, label := range lsLabelQuery {
		f := cache.LabelFilter(label)
		query.Label = append(query.Label, f)
	}

	for _, no := range lsNoQuery {
		switch no {
		case "label":
			query.NoFilters = append(query.NoFilters, cache.NoLabelFilter())
		default:
			return nil, fmt.Errorf("unknown \"no\" filter %s", no)
		}
	}

	switch lsSortBy {
	case "id":
		query.OrderBy = cache.OrderById
	case "creation":
		query.OrderBy = cache.OrderByCreation
	case "edit":
		query.OrderBy = cache.OrderByEdit
	default:
		return nil, fmt.Errorf("unknown sort flag %s", lsSortBy)
	}

	switch lsSortDirection {
	case "asc":
		query.OrderDirection = cache.OrderAscending
	case "desc":
		query.OrderDirection = cache.OrderDescending
	default:
		return nil, fmt.Errorf("unknown sort direction %s", lsSortDirection)
	}

	return query, nil
}

var lsCmd = &cobra.Command{
	Use:   "ls [<query>]",
	Short: "List tickets.",
	Long: `Display a summary of each ticket.

You can pass an additional query to filter and order the list. This query can be expressed either with a simple query language or with flags.`,
	Example: `List vetted tickets sorted by last edition with a query:
git ticket ls status:vetted sort:edit-desc

List merged tickets sorted by creation with flags:
git ticket ls --status merged --by creation
`,
	PreRunE: loadRepo,
	RunE:    runLsBug,
}

func init() {
	RootCmd.AddCommand(lsCmd)

	lsCmd.Flags().SortFlags = false

	lsCmd.Flags().StringSliceVarP(&lsStatusQuery, "status", "s", nil,
		"Filter by status")
	lsCmd.Flags().StringSliceVarP(&lsAuthorQuery, "author", "a", nil,
		"Filter by author")
	lsCmd.Flags().StringSliceVarP(&lsParticipantQuery, "participant", "p", nil,
		"Filter by participant")
	lsCmd.Flags().StringSliceVarP(&lsActorQuery, "actor", "", nil,
		"Filter by actor")
	lsCmd.Flags().StringSliceVarP(&lsAssigneeQuery, "assignee", "A", nil,
		"Filter by assignee")
	lsCmd.Flags().StringSliceVarP(&lsLabelQuery, "label", "l", nil,
		"Filter by label")
	lsCmd.Flags().StringSliceVarP(&lsTitleQuery, "title", "t", nil,
		"Filter by title")
	lsCmd.Flags().StringSliceVarP(&lsNoQuery, "no", "n", nil,
		"Filter by absence of something. Valid values are [label]")
	lsCmd.Flags().StringVarP(&lsSortBy, "by", "b", "creation",
		"Sort the results by a characteristic. Valid values are [id,creation,edit]")
	lsCmd.Flags().StringVarP(&lsSortDirection, "direction", "d", "asc",
		"Select the sorting direction. Valid values are [asc,desc]")
}
