package seed

import "fmt"

// Names of the datasets accepted by Up, Down and Reseed.
const (
	DatasetDev = "dev" // sample orders for local development and demos
	DatasetE2E = "e2e" // orders for the automated UI tests in frontend/e2e
)

// A dataset is the order data seeded on top of the shared base (users, categories,
// styles, artworks). Use one dataset per database: Down removes every order of the
// seeded users. Datasets must not share order keys, since image keys derive from them.
type dataset struct {
	orders        []orderSeed
	payments      []paymentSeed
	deliverables  []deliverableSeed
	cancellations []cancellationSeed
	delayRequests []delayRequestSeed
}

func lookupDataset(name string) (dataset, error) {
	switch name {
	case DatasetDev:
		return devDataset, nil
	case DatasetE2E:
		return e2eDataset, nil
	}
	return dataset{}, fmt.Errorf("unknown seed dataset %q (expected %q or %q)", name, DatasetDev, DatasetE2E)
}

// ValidateDataset reports whether name is a known dataset.
func ValidateDataset(name string) error {
	_, err := lookupDataset(name)
	return err
}
