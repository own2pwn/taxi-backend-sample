package service

type meta struct {
	Distance *int `json:"distance"`
	Time     *int `json:"time"`
}

func (m meta) isEmpty() bool {
	if (m.Distance == nil && m.Time == nil) || (*m.Distance == 0 && *m.Time == 0) {
		return true
	}
	return false
}

func newMeta(distance int, time int) meta {
	return meta{
		Distance: &distance,
		Time:     &time,
	}
}
