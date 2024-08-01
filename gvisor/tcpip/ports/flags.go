package ports

// BitFlags is a bitset representation of Flags.
type BitFlags uint32

const (
	// MostRecentFlag represents Flags.MostRecent.
	MostRecentFlag BitFlags = 1 << iota

	// LoadBalancedFlag represents Flags.LoadBalanced.
	LoadBalancedFlag

	// TupleOnlyFlag represents Flags.TupleOnly.
	TupleOnlyFlag

	// nextFlag is the value that the next added flag will have.
	//
	// It is used to calculate FlagMask below. It is also the number of
	// valid flag states.
	nextFlag

	// FlagMask is a bit mask for BitFlags.
	FlagMask = nextFlag - 1

	// MultiBindFlagMask contains the flags that allow binding the same
	// tuple multiple times.
	MultiBindFlagMask = MostRecentFlag | LoadBalancedFlag
)

// FlagCounter counts how many references each flag combination has.
type FlagCounter struct {
	// refs stores the count for each possible flag combination, (0 though
	// FlagMask).
	refs [nextFlag]int
}
