package httpHandler

import (
	"net/http"

	"github.com/Ignaciotema/YTMusic-web-socket/internal/dispatcher"
)

type HTTPHandler struct {
	dispatcher *dispatcher.Dispatcher
}

func NewHTTPHandler(dispatcher *dispatcher.Dispatcher) *HTTPHandler {
	return &HTTPHandler{
		dispatcher: dispatcher,
	}
}

func (h *HTTPHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	command := r.URL.Query().Get("command")
	if command == "" {
		http.Error(w, "Missing command parameter", http.StatusBadRequest)
		return
	}
	//h.dispatcher.Dispatch(command)
	w.WriteHeader(http.StatusOK)
}
