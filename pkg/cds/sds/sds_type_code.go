package sds

import (
	"encoding/json"
)

type SdsTypeCode string

var sdsTypeCodes = map[int]string{
	0:   "Empty",
	1:   "Object",
	2:   "DBNull",
	3:   "Boolean",
	4:   "Char",
	5:   "SByte",
	6:   "Byte",
	7:   "Int16",
	8:   "UInt16",
	9:   "Int32",
	10:  "UInt32",
	11:  "Int64",
	12:  "UInt64",
	13:  "Single",
	14:  "Double",
	15:  "Decimal",
	16:  "DateTime",
	18:  "String",
	19:  "Guid",
	20:  "DateTimeOffset",
	21:  "TimeSpan",
	22:  "Version",
	103: "NullableBoolean",
	104: "NullableChar",
	105: "NullableSByte",
	106: "NullableByte",
	107: "NullableInt16",
	108: "NullableUInt16",
	109: "NullableInt32",
	110: "NullableUInt32",
	111: "NullableInt64",
	112: "NullableUInt64",
	113: "NullableSingle",
	114: "NullableDouble",
	115: "NullableDecimal",
	116: "NullableDateTime",
	119: "NullableGuid",
	120: "NullableDateTimeOffset",
	121: "NullableTimeSpan",
	203: "BooleanArray",
	204: "CharArray",
	205: "SByteArray",
	206: "ByteArray",
	207: "Int16Array",
	208: "UInt16Array",
	209: "Int32Array",
	210: "UInt32Array",
	211: "Int64Array",
	212: "UInt64Array",
	213: "SingleArray",
	214: "DoubleArray",
	215: "DecimalArray",
	216: "DateTimeArray",
	218: "StringArray",
	219: "GuidArray",
	220: "DateTimeOffsetArray",
	221: "TimeSpanArray",
	222: "VersionArray",
	400: "Array",
	401: "IList",
	402: "IDictionary",
	403: "IEnumerable",
	501: "SdsType",
	502: "SdsTypeProperty",
	605: "SByteEnum",
	606: "ByteEnum",
	607: "Int16Enum",
	608: "UInt16Enum",
	609: "Int32Enum",
	610: "UInt32Enum",
	611: "Int64Enum",
	612: "UInt64Enum",
	705: "NullableSByteEnum",
	706: "NullableByteEnum",
	707: "NullableInt16Enum",
	708: "NullableUInt16Enum",
	709: "NullableInt32Enum",
	710: "NullableUInt32Enum",
	711: "NullableInt64Enum",
	712: "NullableUInt64Enum",
}

func (sdsTypeCode *SdsTypeCode) UnmarshalJSON(b []byte) error {
	var result interface{}
	if err := json.Unmarshal(b, &result); err != nil {
		return err
	}

	switch t := result.(type) {
	case string:
		*sdsTypeCode = SdsTypeCode(t)
	case float64:
		*sdsTypeCode = SdsTypeCode(sdsTypeCodes[int(t)])
	}

	return nil
}
