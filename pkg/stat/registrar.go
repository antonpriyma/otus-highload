package stat

import (
	"reflect"
	"strconv"
	"strings"

	"github.com/antonpriyma/otus-highload/pkg/errors"
	"github.com/iancoleman/strcase"
)

type Registrar interface {
	MustRegister(in interface{})
}

type registrar struct {
	reg Registry
}

func NewRegistrar(reg Registry) Registrar {
	return registrar{reg: reg}
}

func (r registrar) MustRegister(in interface{}) {
	val := reflect.ValueOf(in)
	if val.Kind() != reflect.Ptr {
		panic("input value is not a pointer")
	}

	val = val.Elem()
	if val.Kind() != reflect.Struct {
		panic("input value is not a pointer to struct")
	}

	r.registerStruct(val)
}

func (r registrar) registerStruct(in reflect.Value) {
	for i := 0; i < in.NumField(); i++ {
		field := in.Field(i)
		if field.Kind() == reflect.Struct {
			r.registerStruct(field)
			continue
		}

		field.Set(reflect.ValueOf(r.getMetricForField(field, in.Type().Field(i).Name, in.Type().Field(i).Tag)))
	}
}

var (
	counterType   = reflect.TypeOf((*CounterCtor)(nil)).Elem()
	gaugeType     = reflect.TypeOf((*GaugeCtor)(nil)).Elem()
	histogramType = reflect.TypeOf((*HistogramCtor)(nil)).Elem()
	summaryType   = reflect.TypeOf((*SummaryCtor)(nil)).Elem()
	timerType     = reflect.TypeOf((*TimerCtor)(nil)).Elem()
)

func (r registrar) getMetricForField(val reflect.Value, name string, tag reflect.StructTag) interface{} {
	name = strcase.ToSnake(name)

	switch val.Type() {
	case counterType:
		return r.counterForField(name, tag)
	case gaugeType:
		return r.gaugeForField(name, tag)
	case histogramType:
		return r.histogramForField(name, tag)
	case summaryType:
		return r.summaryForField(name, tag)
	case timerType:
		return r.timerForField(name, tag)
	default:
		panic(errors.Errorf("unknown type: %s", val.Type().Name()))
	}
}

func (r registrar) counterForField(name string, tag reflect.StructTag) CounterCtor {
	labels, constLabels := tagLabels(tag)
	return r.reg.Counter(name, labels, constLabels)
}

func (r registrar) gaugeForField(name string, tag reflect.StructTag) GaugeCtor {
	labels, constLabels := tagLabels(tag)
	return r.reg.Gauge(name, labels, constLabels)
}

func (r registrar) histogramForField(name string, tag reflect.StructTag) HistogramCtor {
	labels, constLabels := tagLabels(tag)
	return r.reg.Histogram(name, tagBuckets(tag), labels, constLabels)
}

func (r registrar) summaryForField(name string, tag reflect.StructTag) SummaryCtor {
	labels, constLabels := tagLabels(tag)
	return r.reg.Summary(name, tagQuantiles(tag), labels, constLabels)
}

func (r registrar) timerForField(name string, tag reflect.StructTag) TimerCtor {
	labels, constLabels := tagLabels(tag)
	return r.reg.Timer(name, timerBuckets(tag), labels, constLabels)
}

const (
	tagNameLabels    = "labels"
	tagNameBuckets   = "buckets"
	tagNameQuantiles = "quantiles"
)

func tagLabels(tag reflect.StructTag) ([]string, Labels) {
	var labelNames []string
	constLabels := Labels{}

	vals := tagStringSlice(tag, tagNameLabels)
	for _, val := range vals {
		split := strings.Split(val, ":")
		switch len(split) {
		case 1:
			labelNames = append(labelNames, val)
		case 2:
			constLabels[split[0]] = split[1]
		default:
			panic(errors.Errorf("unexpected amount of %q in %s", ":", val))
		}
	}

	return labelNames, constLabels
}

var (
	standardTimerBuckets = []float64{
		.001, .002, .003, .004, .005,
		.01, .02, .03, .04, .05,
		.1, .2, .3, .4, .5,
		1, 2, 3, 4, 5,
		10, 20, 30, 40, 50,
	}
)

func timerBuckets(tag reflect.StructTag) []float64 {
	buckets := tagBuckets(tag)
	if len(buckets) == 0 {
		return standardTimerBuckets
	}

	return buckets
}

func tagBuckets(tag reflect.StructTag) []float64 {
	return parseFloat64Slice(tagStringSlice(tag, tagNameBuckets))
}

func tagQuantiles(tag reflect.StructTag) []float64 {
	return parseFloat64Slice(tagStringSlice(tag, tagNameQuantiles))
}

func tagStringSlice(tag reflect.StructTag, name string) []string {
	labelsRaw, ok := tag.Lookup(name)
	if !ok {
		return nil
	}

	return strings.Split(labelsRaw, ",")
}

func parseFloat64Slice(stringSlice []string) []float64 {
	res := make([]float64, 0, len(stringSlice))
	for _, val := range stringSlice {
		resVal, err := strconv.ParseFloat(val, 64)
		if err != nil {
			panic(errors.Wrapf(err, "failed to parse float from %q", val))
		}

		res = append(res, resVal)
	}

	return res
}
