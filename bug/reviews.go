package bug

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/thought-machine/gonduit/requests"

	"github.com/daedaleanai/git-ticket/repository"
)

// PhapTransaction holds common transaction data
type PhabTransaction struct {
	Id        string
	User      string
	Timestamp int64
}

// ReviewComment extends PhabTransaction with comment specific fields
type ReviewComment struct {
	PhabTransaction
	Diff int    // diff id comment was made againt, inline comments only
	Path string // file path, inline comments only
	Line int    // line number, inline comments only
	Text string
}

// ReviewStatus extends PhabTransaction with status specific fields
type ReviewStatus struct {
	PhabTransaction
	Status string
}

// ReviewInfo holds a set of comment and status updates related to a diff
type ReviewInfo struct {
	RevisionId      string // e.g. D1234
	LastTransaction string
	Statuses        []ReviewStatus
	Comments        []ReviewComment
}

// OneLineComment returns a string containing the comment text, and it's an inline
// comment the file & line details, on a single line. Comments over 50 characters
// are truncated.
func (c ReviewComment) OneLineComment() string {
	var output string

	// Put the comment on one line and output the first 50 characters
	oneLineText := strings.ReplaceAll(c.Text, "\n", " ")
	if len(oneLineText) > 50 {
		output = fmt.Sprintf("%.47s...", oneLineText)
	} else {
		output = fmt.Sprintf("%-50s", oneLineText)
	}

	// If it's an inline comment append the file and line number
	if c.Path != "" {
		output = output + fmt.Sprintf(" [%s:%d@%d]", c.Path, c.Line, c.Diff)
	}

	return output
}

// LatestUserStatuses returns a map of users and the latest status they set for
// this review.
func (r ReviewInfo) LatestUserStatuses() map[string]ReviewStatus {
	// Create a map of the latest status change made by all users
	userStatusChange := make(map[string]ReviewStatus)

	for _, s := range r.Statuses {
		if sc, present := userStatusChange[s.User]; !present || s.Timestamp > sc.Timestamp {
			userStatusChange[s.User] = s
		}
	}

	return userStatusChange
}

// FetchReviewInfo exports review comments and status info from Phabricator for
// the given differential ID and returns in a ReviewInfo struct. If a since
// transaction ID is specified then only updates since then are returned.
func FetchReviewInfo(id string, since string) (*ReviewInfo, error) {

	if matched, _ := regexp.MatchString(`^D\d+$`, id); !matched {
		return nil, fmt.Errorf("differential id '%s' unexpected format (Dnnn)", id)
	}

	result := ReviewInfo{RevisionId: id}

	phabClient, err := repository.GetPhabClient()
	if err != nil {
		return nil, err
	}

	var before string
	var after string
	var deltaUpdate bool

	// If since is set then only get the transactions since then, else get them all
	if since != "" {
		before = since
		deltaUpdate = true
	}

	// TODO remove this debug code when review is fully implemented
	//dumpFile, _ := os.Create("dump.json")
	//defer dumpFile.Close()
	//count := 1

	for {

		request := requests.TransactionSearchRequest{ObjectID: id,
			Before: before,
			After:  after,
			Limit:  100}

		// TODO
		//requestJ, _ := json.Marshal(request)
		//dumpFile.WriteString(fmt.Sprintf("\n{\"_comment\":\"=========REQUEST %d==========\"}", count))
		//dumpFile.Write(requestJ)

		response, err := phabClient.TransactionSearch(request)
		if err != nil {
			return nil, err
		}

		// TODO
		//responseJ, _ := json.Marshal(response)
		//dumpFile.WriteString(fmt.Sprintf("\n{\"_comment\":\"=========RESPONSE %d==========\"}", count))
		//dumpFile.Write(responseJ)
		//count++

		if len(response.Data) == 0 {
			break
		}

		// If the Cursor.Before field is blank this response includes the latest
		// transactions, position 0 has the newest
		if response.Cursor.Before == nil {
			result.LastTransaction = strconv.Itoa(response.Data[0].ID)
		}

		// Loop through all transactions
		for _, t := range response.Data {

			transData := PhabTransaction{Id: strconv.Itoa(t.ID),
				User:      t.AuthorPHID,
				Timestamp: time.Time(t.DateCreated).Unix()}

			switch t.Type {
			case "inline":
				// If it's an inline comment the Fields contains the file path, line and diff ID
				diff := t.Fields["diff"].(map[string]interface{})
				commentDiff := int(diff["id"].(float64))
				commentPath := t.Fields["path"].(string)
				commentLine := int(t.Fields["line"].(float64))

				for _, c := range t.Comments {
					newComment := ReviewComment{PhabTransaction: transData,
						Diff: commentDiff,
						Path: commentPath,
						Line: commentLine,
						Text: c.Content["raw"].(string)}

					result.Comments = append(result.Comments, newComment)
				}

			case "comment":
				for _, c := range t.Comments {
					newComment := ReviewComment{PhabTransaction: transData,
						Text: c.Content["raw"].(string)}

					result.Comments = append(result.Comments, newComment)
				}

			case "status":
				newStatus := ReviewStatus{PhabTransaction: transData,
					Status: t.Fields["new"].(string)}

				result.Statuses = append(result.Statuses, newStatus)

			case "accept":
				newStatus := ReviewStatus{PhabTransaction: transData,
					Status: "accepted"}

				result.Statuses = append(result.Statuses, newStatus)

			}
		}

		if deltaUpdate {
			// If we requested only transactions after a certain one (by setting the request
			// "before" field) then Phabricator sends the oldest transactions first, if there's
			// more than the "limit" remaining then the Cursor.Before field will be set to
			// indicate more newer ones are available.
			if response.Cursor.Before == nil {
				// there's no more transactions to get
				break
			}
			before = response.Cursor.Before.(string)
		} else {
			// If we requested all transactions then Phabricator sends the newest transactions
			// first, if there's more than the "limit" remaining then the Cursor.After field
			// will be set to indicate more older ones are available.
			if response.Cursor.After == nil {
				// there's no more transactions to get
				break
			}
			after = response.Cursor.After.(string)
		}

	}

	return &result, nil
}
