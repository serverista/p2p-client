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

	// ListUserServicesEndpoint
	ListUserServicesEndpoint = Endpoint{
		Method: "GET",
		Uri:    "/v1/services",
	}

	// GetUserServiceEndpoint
	GetUserServiceEndpoint = Endpoint{
		Method: "GET",
		Uri:    "/v1/services/%d",
	}

	// CreateServicesEndpoint
	CreateServicesEndpoint = Endpoint{
		Method: "POST",
		Uri:    "/v1/services",
	}

	// ManageServiceEndpoint
	ManageServiceEndpoint = Endpoint{
		Method: "POST",
		Uri:    "/v1/services/%d/management",
	}
)
