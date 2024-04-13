package deleter

import (
	"log/slog"
	"net/http"
	"strconv"

	"github.com/P1coFly/CarInfoEM/http-server/handlers/err_response"
	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
	"github.com/go-chi/render"
)

type Request struct {
	CarID int `json:"car_id,omitempty"`
}

type DeleterCar interface {
	DeleteCar(carID int) (int, error)
}

// @Summary Delete
// @Tags car
// @Description delete car
// @Accept json
// @Produce json
// @Param id path int true "Car ID"
// @Success 204
// @Failure 400,404 {object} err_response.Response
// @Failure 500 {object} err_response.Response
// @Failure default {object} err_response.Response
// @Router /car/delete/{id} [delete]
func New(log *slog.Logger, deleter DeleterCar) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const op = "handlers.DeleterCar.New"

		log = log.With(
			slog.String("op", op),
			slog.String("request_id", middleware.GetReqID(r.Context())),
		)

		//пытаемсяя получить id с запроса
		idx, err := strconv.Atoi(chi.URLParam(r, "id"))

		if err != nil {
			log.Error("failed to get id")
			w.WriteHeader(400)
			render.JSON(w, r, err_response.Error("failed to get. Need car's id"))

			return
		}

		req := Request{CarID: idx}

		log.Debug("carID", slog.Any("carID", req.CarID))
		log.Info("request body decoded", slog.Any("request", req))

		//удаляем машину
		id, err := deleter.DeleteCar(req.CarID)
		if err != nil {
			log.Error("failed delete car", "error", err)
			if id == -1 {
				w.WriteHeader(404)
				render.JSON(w, r, err_response.Error("car with this id was not found"))
				return
			}
			w.WriteHeader(500)
			render.JSON(w, r, err_response.Error("failed delete car"))

			return
		}

		log.Info("car deleted", slog.Int("id", req.CarID))

		w.WriteHeader(204)
	}
}
