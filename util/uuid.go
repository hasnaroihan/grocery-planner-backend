package util

import "github.com/google/uuid"

func ConvertUUIDString(uuid_string string) (uuid.UUID, error) {
	id, err := uuid.Parse(uuid_string)
	if err != nil {
		return uuid.Nil, err
	}

	return id, nil
}