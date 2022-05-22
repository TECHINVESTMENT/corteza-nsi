package test

import (
	"context"
	"fmt"
	"strconv"
	"testing"
	"time"

	"github.com/cortezaproject/corteza-server/compose/types"
	"github.com/cortezaproject/corteza-server/pkg/dal"
	"github.com/cortezaproject/corteza-server/pkg/filter"
	"github.com/cortezaproject/corteza-server/pkg/logger"
	"github.com/stretchr/testify/require"
)

func All(t *testing.T, d dal.Connection) {
	t.Run("RecordCodec", func(t *testing.T) { RecordCodec(t, d) })
	t.Run("RecordSearch", func(t *testing.T) { RecordSearch(t, d) })
}

func RecordCodec(t *testing.T, d dal.Connection) {
	var (
		req = require.New(t)

		// enable query logging when +debug is used on DSN schema
		ctx = logger.ContextWithValue(context.Background(), logger.MakeDebugLogger())

		m = &dal.Model{
			Ident: "crs_test_codec",
			Attributes: dal.AttributeSet{
				&dal.Attribute{Ident: "ID", Type: &dal.TypeID{}, Store: &dal.CodecAlias{Ident: "id"}, PrimaryKey: true},
				&dal.Attribute{Ident: "createdAt", Type: &dal.TypeTimestamp{}, Store: &dal.CodecAlias{Ident: "created_at"}},
				&dal.Attribute{Ident: "updatedAt", Type: &dal.TypeTimestamp{}, Store: &dal.CodecAlias{Ident: "updated_at"}},

				&dal.Attribute{Ident: "vID", Type: &dal.TypeID{}, Store: &dal.CodecRecordValueSetJSON{Ident: "meta"}},
				&dal.Attribute{Ident: "vRef", Type: &dal.TypeRef{}, Store: &dal.CodecRecordValueSetJSON{Ident: "meta"}},
				&dal.Attribute{Ident: "vTimestamp", Type: &dal.TypeTimestamp{}, Store: &dal.CodecRecordValueSetJSON{Ident: "meta"}},
				&dal.Attribute{Ident: "vTime", Type: &dal.TypeTime{}, Store: &dal.CodecRecordValueSetJSON{Ident: "meta"}},
				&dal.Attribute{Ident: "vDate", Type: &dal.TypeDate{}, Store: &dal.CodecRecordValueSetJSON{Ident: "meta"}},
				&dal.Attribute{Ident: "vNumber", Type: &dal.TypeNumber{}, Store: &dal.CodecRecordValueSetJSON{Ident: "meta"}},
				&dal.Attribute{Ident: "vText", Type: &dal.TypeText{}, Store: &dal.CodecRecordValueSetJSON{Ident: "meta"}},
				&dal.Attribute{Ident: "vBoolean_T", Type: &dal.TypeBoolean{}, Store: &dal.CodecRecordValueSetJSON{Ident: "meta"}},
				&dal.Attribute{Ident: "vBoolean_F", Type: &dal.TypeBoolean{}, Store: &dal.CodecRecordValueSetJSON{Ident: "meta"}},
				&dal.Attribute{Ident: "vEnum", Type: &dal.TypeEnum{}, Store: &dal.CodecRecordValueSetJSON{Ident: "meta"}},
				&dal.Attribute{Ident: "vGeometry", Type: &dal.TypeGeometry{}, Store: &dal.CodecRecordValueSetJSON{Ident: "meta"}},
				&dal.Attribute{Ident: "vJSON", Type: &dal.TypeJSON{}, Store: &dal.CodecRecordValueSetJSON{Ident: "meta"}},
				&dal.Attribute{Ident: "vBlob", Type: &dal.TypeBlob{}, Store: &dal.CodecRecordValueSetJSON{Ident: "meta"}},
				&dal.Attribute{Ident: "vUUID", Type: &dal.TypeUUID{}, Store: &dal.CodecRecordValueSetJSON{Ident: "meta"}},
				&dal.Attribute{Ident: "pID", Type: &dal.TypeID{}, Store: &dal.CodecPlain{}},
				&dal.Attribute{Ident: "pRef", Type: &dal.TypeRef{}, Store: &dal.CodecPlain{}},
				&dal.Attribute{Ident: "pTimestamp_TZT", Type: &dal.TypeTimestamp{Timezone: true}, Store: &dal.CodecPlain{}},
				&dal.Attribute{Ident: "pTimestamp_TZF", Type: &dal.TypeTimestamp{Timezone: false}, Store: &dal.CodecPlain{}},
				&dal.Attribute{Ident: "pTime", Type: &dal.TypeTime{}, Store: &dal.CodecPlain{}},
				&dal.Attribute{Ident: "pDate", Type: &dal.TypeDate{}, Store: &dal.CodecPlain{}},
				&dal.Attribute{Ident: "pNumber", Type: &dal.TypeNumber{}, Store: &dal.CodecPlain{}},
				&dal.Attribute{Ident: "pText", Type: &dal.TypeText{}, Store: &dal.CodecPlain{}},
				&dal.Attribute{Ident: "pBoolean_T", Type: &dal.TypeBoolean{}, Store: &dal.CodecPlain{}},
				&dal.Attribute{Ident: "pBoolean_F", Type: &dal.TypeBoolean{}, Store: &dal.CodecPlain{}},
				&dal.Attribute{Ident: "pEnum", Type: &dal.TypeEnum{}, Store: &dal.CodecPlain{}},
				&dal.Attribute{Ident: "pGeometry", Type: &dal.TypeGeometry{}, Store: &dal.CodecPlain{}},
				&dal.Attribute{Ident: "pJSON", Type: &dal.TypeJSON{}, Store: &dal.CodecPlain{}},
				&dal.Attribute{Ident: "pBlob", Type: &dal.TypeBlob{}, Store: &dal.CodecPlain{}},
				&dal.Attribute{Ident: "pUUID", Type: &dal.TypeUUID{}, Store: &dal.CodecPlain{}},
			},
		}

		rIn  = types.Record{ID: 42}
		err  error
		rOut *types.Record

		piTime time.Time
	)

	piTime, err = time.Parse("2006-01-02T15:04:05", "2006-01-02T15:04:05")
	req.NoError(err)
	piTime = piTime.UTC()

	rIn.CreatedAt = piTime
	rIn.UpdatedAt = &piTime

	rIn.Values = rIn.Values.Set(&types.RecordValue{Name: "vID", Value: "34324"})
	rIn.Values = rIn.Values.Set(&types.RecordValue{Name: "vRef", Value: "32423"})
	rIn.Values = rIn.Values.Set(&types.RecordValue{Name: "vTimestamp", Value: "2022-01-01T10:10:10+02:00"})
	rIn.Values = rIn.Values.Set(&types.RecordValue{Name: "vTime", Value: "04:10:10+04:00"})
	rIn.Values = rIn.Values.Set(&types.RecordValue{Name: "vDate", Value: "2022-01-01"})
	rIn.Values = rIn.Values.Set(&types.RecordValue{Name: "vNumber", Value: "2423423"})
	rIn.Values = rIn.Values.Set(&types.RecordValue{Name: "vText", Value: "lorm ipsum "})
	rIn.Values = rIn.Values.Set(&types.RecordValue{Name: "vBoolean_T", Value: "true"})
	rIn.Values = rIn.Values.Set(&types.RecordValue{Name: "vBoolean_F", Value: "false"})
	rIn.Values = rIn.Values.Set(&types.RecordValue{Name: "vEnum", Value: "abc"})
	rIn.Values = rIn.Values.Set(&types.RecordValue{Name: "vGeometry", Value: `{"lat":1,"lng":1}`})
	rIn.Values = rIn.Values.Set(&types.RecordValue{Name: "vJSON", Value: `[{"bool":true"}]`})
	rIn.Values = rIn.Values.Set(&types.RecordValue{Name: "vBlob", Value: "0110101"})
	rIn.Values = rIn.Values.Set(&types.RecordValue{Name: "vUUID", Value: "ba485865-54f9-44de-bde8-6965556c022a"})
	rIn.Values = rIn.Values.Set(&types.RecordValue{Name: "pID", Value: "34324"})
	rIn.Values = rIn.Values.Set(&types.RecordValue{Name: "pRef", Value: "32423"})
	rIn.Values = rIn.Values.Set(&types.RecordValue{Name: "pTimestamp_TZF", Value: "2022-02-01T10:10:10"})

	// @todo how (if at all) should we know if underlying DB supports timezone?
	//rIn.Values = rIn.Values.Set(&types.RecordValue{Name: "pTimestamp_TZT", Value: "2022-02-01T10:10:10"})

	rIn.Values = rIn.Values.Set(&types.RecordValue{Name: "pTime", Value: "06:06:06"})
	rIn.Values = rIn.Values.Set(&types.RecordValue{Name: "pDate", Value: "2022-01-01"})
	rIn.Values = rIn.Values.Set(&types.RecordValue{Name: "pNumber", Value: "2423423"})
	rIn.Values = rIn.Values.Set(&types.RecordValue{Name: "pText", Value: "lorm ipsum "})
	rIn.Values = rIn.Values.Set(&types.RecordValue{Name: "pBoolean_T", Value: "true"})
	rIn.Values = rIn.Values.Set(&types.RecordValue{Name: "pBoolean_F", Value: "false"})
	rIn.Values = rIn.Values.Set(&types.RecordValue{Name: "pEnum", Value: "abc"})
	rIn.Values = rIn.Values.Set(&types.RecordValue{Name: "pGeometry", Value: `{"lat":1,"lng":1}`})
	rIn.Values = rIn.Values.Set(&types.RecordValue{Name: "pJSON", Value: `[{"bool":true"}]`})
	rIn.Values = rIn.Values.Set(&types.RecordValue{Name: "pBlob", Value: "0110101"})
	rIn.Values = rIn.Values.Set(&types.RecordValue{Name: "pUUID", Value: "ba485865-54f9-44de-bde8-6965556c022a"})
	rIn.Values = rIn.Values.GetClean()

	req.NoError(d.Create(ctx, m, &rIn))

	rOut = new(types.Record)
	req.NoError(d.Lookup(ctx, m, dal.PKValues{"id": rIn.ID}, rOut))

	{
		// normalize timezone on timestamps
		rOut.CreatedAt = rOut.CreatedAt.UTC()
		aux := rOut.UpdatedAt.UTC()
		rOut.UpdatedAt = &aux
	}

	for _, attr := range m.Attributes {
		vIn, err := rIn.GetValue(attr.Ident, 0)
		req.NoError(err)
		vOut, err := rOut.GetValue(attr.Ident, 0)
		req.NoError(err)
		req.Equal(vIn, vOut, "values for attribute %q are not equal", attr.Ident)
	}
}

func RecordSearch(t *testing.T, d dal.Connection) {
	const (
		totalRecords = 10
	)

	var (
		req = require.New(t)

		// enable query logging when +debug is used on DSN schema
		ctx = logger.ContextWithValue(context.Background(), logger.MakeDebugLogger())

		m = &dal.Model{
			Ident: "crs_test_search",
			Attributes: dal.AttributeSet{
				&dal.Attribute{Ident: "ID", Type: &dal.TypeID{}, Store: &dal.CodecAlias{Ident: "id"}, PrimaryKey: true},
				&dal.Attribute{Ident: "createdAt", Type: &dal.TypeTimestamp{}, Store: &dal.CodecAlias{Ident: "created_at"}},
				&dal.Attribute{Ident: "updatedAt", Type: &dal.TypeTimestamp{}, Store: &dal.CodecAlias{Ident: "updated_at"}},

				&dal.Attribute{Ident: "v_string", Type: &dal.TypeText{}, Store: &dal.CodecRecordValueSetJSON{Ident: "meta"}},
				&dal.Attribute{Ident: "v_number", Type: &dal.TypeNumber{}, Store: &dal.CodecRecordValueSetJSON{Ident: "meta"}},
				&dal.Attribute{Ident: "v_is_odd", Type: &dal.TypeBoolean{}, Store: &dal.CodecRecordValueSetJSON{Ident: "meta"}},
				&dal.Attribute{Ident: "p_string", Type: &dal.TypeText{}, Store: &dal.CodecPlain{}},
				&dal.Attribute{Ident: "p_number", Type: &dal.TypeNumber{}, Store: &dal.CodecPlain{}},
				&dal.Attribute{Ident: "p_is_odd", Type: &dal.TypeBoolean{}, Store: &dal.CodecPlain{}},
			},
		}
	)

	for ID := uint64(1); ID <= totalRecords; ID++ {
		r := &types.Record{ID: ID, CreatedAt: time.Now()}

		i := int(ID)
		r.Values = r.Values.Set(&types.RecordValue{Name: "v_string", Value: "tens_" + strconv.Itoa(i%10)})
		r.Values = r.Values.Set(&types.RecordValue{Name: "v_number", Value: strconv.Itoa(i)})
		r.Values = r.Values.Set(&types.RecordValue{Name: "v_is_odd", Value: strconv.FormatBool(i%2 == 1)})
		r.Values = r.Values.Set(&types.RecordValue{Name: "p_string", Value: "tens_" + strconv.Itoa(i%10)})
		r.Values = r.Values.Set(&types.RecordValue{Name: "p_number", Value: strconv.Itoa(i)})
		r.Values = r.Values.Set(&types.RecordValue{Name: "p_is_odd", Value: strconv.FormatBool(i%2 == 1)})

		req.NoError(d.Create(ctx, m, r))
	}

	cases := []struct {
		f     types.RecordFilter
		total int
	}{
		{
			total: totalRecords,
		},
		{
			f:     types.RecordFilter{Query: "v_string == p_string"},
			total: totalRecords,
		},
		{
			f:     types.RecordFilter{Query: "v_number == p_number"},
			total: totalRecords,
		},
		{
			f:     types.RecordFilter{Query: "p_is_odd"},
			total: totalRecords / 2,
		},
		{
			f:     types.RecordFilter{Query: "true = p_is_odd"},
			total: totalRecords / 2,
		},
		{
			f:     types.RecordFilter{Query: "p_is_odd = true"},
			total: totalRecords / 2,
		},
		{
			f:     types.RecordFilter{Query: "!p_is_odd"},
			total: totalRecords / 2,
		},
		{
			f:     types.RecordFilter{Query: "p_number = 1"},
			total: 1,
		},
	}

	for _, c := range cases {
		t.Run(c.f.Query, func(t *testing.T) {
			var (
				req = require.New(t)
			)

			i, err := d.Search(ctx, m, c.f.ToFilter())
			req.NoError(err)

			rr, err := drain(ctx, i)
			req.NoError(err)
			req.Len(rr, c.total)
		})
	}

	t.Run("paging", func(t *testing.T) {
		var (
			req      = require.New(t)
			ids      string
			fwd, bck *filter.PagingCursor

			search = func(where, orderBy string, lim uint, cur *filter.PagingCursor) (ids string, fwd, bck *filter.PagingCursor) {
				f := types.RecordFilter{Query: where}
				f.PageCursor = cur
				f.Limit = lim
				req.NoError(f.Sort.Set(orderBy))
				i, err := d.Search(ctx, m, f.ToFilter())
				req.NoError(err)
				req.NoError(i.Err())

				if !i.Next(ctx) {
					req.NoError(i.Err())
					return
				}

				r := new(types.Record)
				req.NoError(i.Scan(r))

				bck, err = i.BackCursor(r)
				req.NoError(err)
				t.Logf("bck-cursor (from the 1st fetched record): %v", bck)

				rr, err := drain(ctx, i)
				req.NoError(err)

				if len(rr) > 0 {
					fwd, err = i.ForwardCursor(rr[len(rr)-1])
					req.NoError(err)
					t.Logf("fwd-cursor (from the lst fetched record): %v", fwd)
				}

				ids = fmt.Sprintf("%d", r.ID)
				for _, r = range rr {
					ids += fmt.Sprintf(",%d", r.ID)
				}

				return
			}
		)

		ids, fwd, _ = search("", "", 3, nil)
		req.Equal("1,2,3", ids)
		ids, fwd, _ = search("", "", 3, fwd)
		req.Equal("4,5,6", ids)
		ids, fwd, _ = search("", "", 3, fwd)
		req.Equal("7,8,9", ids)
		ids, _, bck = search("", "", 3, fwd)
		req.Equal("10", ids)

		ids, _, bck = search("", "", 3, bck)
		req.Equal("7,8,9", ids)
		ids, _, bck = search("", "", 3, bck)
		req.Equal("4,5,6", ids)
		ids, _, bck = search("", "", 3, bck)
		req.Equal("1,2,3", ids)

		ids, fwd, _ = search("", "p_is_odd", 3, nil)
		req.Equal("2,4,6", ids)
		ids, _, bck = search("", "", 3, fwd)
		req.Equal("8,10,1", ids)
		ids, _, _ = search("", "", 3, bck)
		req.Equal("2,4,6", ids)

		ids, fwd, _ = search("", "v_is_odd", 3, nil)
		req.Equal("2,4,6", ids)
		ids, _, bck = search("", "", 3, fwd)
		req.Equal("8,10,1", ids)
		ids, _, _ = search("", "", 3, bck)
		req.Equal("2,4,6", ids)

		_, _ = fwd, bck // avoiding unused var. error
	})
}

func drain(ctx context.Context, i dal.Iterator) (rr []*types.Record, err error) {
	var r *types.Record
	rr = make([]*types.Record, 0, 100)
	for i.Next(ctx) {
		if i.Err() != nil {
			return nil, i.Err()
		}

		r = new(types.Record)
		if err = i.Scan(r); err != nil {
			return
		}

		rr = append(rr, r)
	}

	return rr, i.Err()
}