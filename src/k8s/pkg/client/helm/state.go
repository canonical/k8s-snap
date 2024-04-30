package helm

// State is used to define how Client.Apply() handles install, upgrade or delete operations.
type State int

const (
	// StateDeleted means that the chart should not be installed.
	StateDeleted State = iota

	// StatePresent means that the chart must be present. If it already exists, it is upgraded with the new configuration, otherwise it is installed.
	StatePresent

	// StateUpgradeOnly means that the chart will be refreshed if installed, fail otherwise.
	StateUpgradeOnly
)

func StatePresentOrDeleted(enabled bool) State {
	if enabled {
		return StatePresent
	}
	return StateDeleted
}

func StateUpgradeOnlyOrDeleted(enabled bool) State {
	if enabled {
		return StateUpgradeOnly
	}
	return StateDeleted
}
