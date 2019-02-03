package service

// Response - струтура, описывающая формат ответа сервиса
type Response struct {
	Result taxiResult `json:"results"`
	Meta   *meta      `json:"meta,omitempty"`
}

type taxiResult struct {
	ID      int          `json:"id"`
	Optimal *resultBlock `json:"optimal,omitempty"`
	Else    *resultBlock `json:"else,omitempty"`
}

type resultBlock struct {
	Title   string          `json:"title"`
	Summary string          `json:"summary"`
	Results []serviceRecord `json:"results"`
}

func newResponse(m *meta, optimal, elses []serviceRecord) *Response {
	var optimalResBlock, elsesResBlock *resultBlock
	if optimal != nil && len(optimal) > 0 {
		optimalResBlock = &resultBlock{
			Summary: "Оптимальный выбор с учётом рейтинга перевозчика и популярности",
			Results: optimal,
		}
	}
	if elses != nil && len(elses) > 0 {
		elsesResBlock = &resultBlock{
			Results: elses,
		}
	}
	return &Response{
		Result: taxiResult{
			ID:      -1,
			Optimal: optimalResBlock,
			Else:    elsesResBlock,
		},
		Meta: m,
	}
}
