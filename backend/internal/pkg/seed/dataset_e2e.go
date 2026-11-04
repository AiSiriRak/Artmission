package seed

// e2eOrder returns an order of customer-1 and artist-2 for the automated UI tests
// in frontend/e2e, due in 14 days. Order keys must not collide with devDataset's
// (see dataset). The tests look these orders up by name, so keep the names stable.
func e2eOrder(key, name, description, status string) orderSeed {
	return orderSeed{
		Key: key, CustomerKey: "customer-1", ArtistKey: "artist-2", ArtworkKey: "artwork-4",
		Name: name, CustomerDescription: description,
		Status: status, DeadlineDays: 14,
	}
}

var e2eDataset = func() dataset {
	completed := e2eOrder("e2e-completed", "E2E completed order", "Finished order, for read-only status checks.", "SUCCESS")
	completed.Completed = true

	return dataset{
		orders: []orderSeed{
			e2eOrder("e2e-postpone-send", "E2E postpone - send request", "Order for the deadline extension request test.", "IN_PROCESS"),
			e2eOrder("e2e-postpone-validation", "E2E postpone - validation", "Order for the deadline extension validation tests.", "IN_PROCESS"),
			e2eOrder("e2e-postpone-pending", "E2E postpone - request pending", "Order that already has an extension request awaiting the customer.", "IN_PROCESS"),
			e2eOrder("e2e-draft-jpg", "E2E draft - JPEG", "Order for submitting a JPEG draft.", "IN_PROCESS"),
			e2eOrder("e2e-draft-png", "E2E draft - PNG", "Order for submitting a PNG draft.", "IN_PROCESS"),
			e2eOrder("e2e-draft-webp", "E2E draft - WebP", "Order for submitting a WebP draft.", "IN_PROCESS"),
			e2eOrder("e2e-draft-validation", "E2E draft - validation", "Order for the draft upload validation tests.", "IN_PROCESS"),
			e2eOrder("e2e-draft-awaiting-review", "E2E draft - awaiting review", "Order whose latest draft waits for the customer.", "IN_PROCESS"),
			e2eOrder("e2e-not-paid", "E2E not paid order", "Accepted order waiting for payment, for read-only status checks.", "NOT_PAID"),
			completed,
			e2eOrder("e2e-cancelled", "E2E cancelled order", "Cancelled order, for read-only status checks.", "CANCEL"),
			// The only order of artist-3 (cleo), so deleting that account is allowed.
			{
				Key: "e2e-artist-3-cancelled", CustomerKey: "customer-1", ArtistKey: "artist-3", ArtworkKey: "artwork-5",
				Name: "E2E cleo's cancelled order", CustomerDescription: "Only order of an artist whose account is deleted by a test.",
				Status: "CANCEL", DeadlineDays: 14,
			},
		},
		payments: []paymentSeed{
			{OrderKey: "e2e-not-paid"},
		},
		deliverables: []deliverableSeed{
			{OrderKey: "e2e-draft-awaiting-review", Version: 1, Decision: "WAIT"},
			{OrderKey: "e2e-completed", Version: 1, Decision: "APPROVED"},
		},
		cancellations: []cancellationSeed{
			{OrderKey: "e2e-cancelled", Reason: "NO_LONGER_NEEDED", Explanation: "Cancelled for the automated tests."},
			{OrderKey: "e2e-artist-3-cancelled", Reason: "NO_LONGER_NEEDED", Explanation: "Cancelled for the automated tests."},
		},
		delayRequests: []delayRequestSeed{
			{
				OrderKey: "e2e-postpone-pending", ExtraDays: 7,
				Reason: "NEED_MORE_TIME", Explanation: "I need more time for details.", Decision: "WAIT",
			},
		},
	}
}()
