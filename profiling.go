package duckdb

/*
#include <duckdb.h>
*/
import "C"

import (
	"unsafe"
)

type ProfilingInfo struct {
	Metrics  map[string]string
	Children []ProfilingInfo
}

func GetProfilingInfo(driverConn any) (ProfilingInfo, error) {
	info := ProfilingInfo{}

	con, ok := driverConn.(*conn)
	if !ok {
		return info, getError(errInvalidCon, nil)
	}
	if con.closed {
		return info, getError(errClosedCon, nil)
	}

	duckdbInfo := C.duckdb_get_profiling_info(con.duckdbCon)
	if duckdbInfo == nil {
		return info, getError(errProfilingInfoEmpty, nil)
	}

	// Recursive tree traversal.
	info.getMetrics(duckdbInfo)
	return info, nil
}

func (info *ProfilingInfo) getMetrics(duckdbInfo C.duckdb_profiling_info) {
	m := C.duckdb_profiling_info_get_metrics(duckdbInfo)
	count := C.duckdb_get_map_size(m)

	for i := C.idx_t(0); i < count; i++ {
		key := C.duckdb_get_map_key(m, i)
		value := C.duckdb_get_map_value(m, i)

		cKey := C.duckdb_get_varchar(key)
		cValue := C.duckdb_get_varchar(value)

		keyStr := C.GoString(cKey)
		valueStr := C.GoString(cValue)

		info.Metrics[keyStr] = valueStr

		C.duckdb_destroy_value(&key)
		C.duckdb_destroy_value(&value)
		C.duckdb_free(unsafe.Pointer(cKey))
		C.duckdb_free(unsafe.Pointer(cValue))
	}
	C.duckdb_destroy_value(&m)

	childCount := C.duckdb_profiling_info_get_child_count(duckdbInfo)
	for i := C.idx_t(0); i < childCount; i++ {
		duckdbChildInfo := C.duckdb_profiling_info_get_child(duckdbInfo, i)
		childInfo := ProfilingInfo{}
		childInfo.getMetrics(duckdbChildInfo)
		info.Children = append(info.Children, childInfo)
	}
}
