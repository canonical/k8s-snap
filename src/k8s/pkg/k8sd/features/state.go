package features

// state is used to define how Manager.Apply() handles install, upgrade or delete operations.
type state int

const (
	// stateDeleted means that the feature should not be installed.
	stateDeleted state = iota

	// statePresent means that the feature must be present. If it already exists, it is upgraded with the new configuration, otherwise it is installed.
	statePresent

	// stateUpgradeOnly means that the feature will be refreshed if installed, fail otherwise.
	stateUpgradeOnly
)

func statePresentOrDeleted(enabled bool) state {
	if enabled {
		return statePresent
	}
	return stateDeleted
}

func stateUpgradeOnlyOrDeleted(enabled bool) state {
	if enabled {
		return stateUpgradeOnly
	}
	return stateDeleted
}
