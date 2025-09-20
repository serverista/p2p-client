package p2pclient

type Endpoint struct {
	Method string
	Uri    string
}

var (
	// PlansEndpoint is a public endpoint
	PlansEndpoint = Endpoint{
		Method: "GET",
		Uri:    "/v1/plans",
	}

	// ListUserServicesEndpoint requires message signing with nonce and timestamp
	ListUserServicesEndpoint = Endpoint{
		Method: "GET",
		Uri:    "/v1/services",
	}

	// CreateServicesEndpoint requires message signing with nonce and timestamp
	CreateServicesEndpoint = Endpoint{
		Method: "POST",
		Uri:    "/v1/services",
	}
)
