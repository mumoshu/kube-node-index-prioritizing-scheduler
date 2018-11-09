package extender

import "github.com/julienschmidt/httprouter"

func NewHandler(predicates []Predicate, priorities []Prioritize) *httprouter.Router {
	router := httprouter.New()
	AddVersion(router)

	for _, p := range predicates {
		AddPredicate(router, p)
	}

	for _, p := range priorities {
		AddPrioritize(router, p)
	}

	AddBind(router, NoBind)

	return router
}
