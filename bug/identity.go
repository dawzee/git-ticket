package bug

import (
	"github.com/daedaleanai/git-ticket/identity"
)

// EnsureIdentities walk the graph of operations and make sure that all Identity
// are properly loaded. That is, it replace all the IdentityStub with the full
// Identity, loaded through a Resolver.
func (bug *Bug) EnsureIdentities(resolver identity.Resolver) error {
	it := NewOperationIterator(bug)

	for it.Next() {
		op := it.Value()
		base := op.base()

		if stub, ok := base.Author.(*identity.IdentityStub); ok {
			id, err := resolver.ResolveIdentity(stub.Id())
			if err != nil {
				return err
			}

			base.Author = id
		}

		// if it's an operation which sets the assignee we also need to
		// resolve that
		if setAss, ok := op.(*SetAssigneeOperation); ok {
			id, err := resolver.ResolveIdentity(setAss.Assignee.Id())
			if err != nil {
				return err
			}

			setAss.Assignee = id
		}
	}
	return nil
}
