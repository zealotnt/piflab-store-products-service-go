package lib

import (
	"errors"
	"net/url"
	"strconv"
	"strings"
)

type Context struct {
	Params    map[string]string
	GetParams url.Values
}

func (c Context) IDs() []uint64 {
	id_param := c.Params["id"]

	// Try to parse Id by 1 number, if success, return right away
	id, err := strconv.ParseUint(id_param, 10, 32)
	if err == nil {
		return []uint64{id}
	}

	// Try to parse Id by comma seperated list
	str_list := strings.Split(id_param, ",")
	var uint_list []uint64
	for _, str := range str_list {
		id, err = strconv.ParseUint(str, 10, 32)
		uint_list = append(uint_list, id)
		if err != nil {
			return []uint64{0}
		}
	}

	return uint_list
}

func (c Context) ID() uint64 {
	id, err := strconv.ParseUint(c.Params["id"], 10, 32)

	if err != nil {
		return 0
	}

	return id
}

func (c Context) Limit() (uint, error) {
	limit, err := strconv.ParseInt(c.Params["limit"], 10, 32)

	if err != nil {
		return 0, errors.New("Error when parsing limit parameter")
	}

	if limit <= 0 {
		return 0, errors.New("Limit must bigger than 0")
	}

	return uint(limit), nil
}

func (c Context) Offset() (uint, error) {
	offset, err := strconv.ParseInt(c.Params["offset"], 10, 32)

	if err != nil {
		return 0, errors.New("Error when parsing offset parameter")
	}

	if offset < 0 {
		return 0, errors.New("Offset must bigger than or equal to 0")
	}

	return uint(offset), nil
}
