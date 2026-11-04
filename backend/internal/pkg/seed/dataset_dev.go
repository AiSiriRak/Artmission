package seed

var devDataset = dataset{
	orders: []orderSeed{
		{
			Key: "order-1", CustomerKey: "customer-1", ArtistKey: "artist-1", ArtworkKey: "artwork-1",
			Name: "Portrait for mom's birthday", CustomerDescription: "Please use a warm color palette.",
			Status: "PENDING",
		},
		{
			Key: "order-2", CustomerKey: "customer-1", ArtistKey: "artist-1", ArtworkKey: "artwork-2",
			Name: "Anime avatar", CustomerDescription: "Blue hair, casual outfit.",
			Status: "NOT_PAID",
		},
		{
			Key: "order-3", CustomerKey: "customer-1", ArtistKey: "artist-2", ArtworkKey: "artwork-3",
			Name: "Dream landscape", CustomerDescription: "A mountain valley at sunrise.",
			Status: "IN_PROCESS",
		},
		{
			Key: "order-4", CustomerKey: "customer-1", ArtistKey: "artist-2", ArtworkKey: "artwork-4",
			Name: "Living room wall art", CustomerDescription: "Calm colors, minimal detail.",
			Status: "SUCCESS", Completed: true,
		},
		{
			Key: "order-5", CustomerKey: "customer-1", ArtistKey: "artist-3", ArtworkKey: "artwork-5",
			Name: "Line art self-portrait", CustomerDescription: "Simple black line art.",
			Status: "CANCEL",
		},
	},
	payments: []paymentSeed{
		{OrderKey: "order-2"},
	},
	deliverables: []deliverableSeed{
		{OrderKey: "order-3", Version: 1, Decision: "WAIT"},
		{OrderKey: "order-4", Version: 1, Decision: "APPROVED"},
	},
	cancellations: []cancellationSeed{
		{OrderKey: "order-5", Reason: "NO_LONGER_NEEDED", Explanation: "I changed my mind about the style."},
	},
}
