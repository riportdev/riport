package jobs

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/riportdev/riport/db/migration/jobs"
	"github.com/riportdev/riport/db/sqlite"
	"github.com/riportdev/riport/server/test/jb"
	"github.com/riportdev/riport/share/models"
	"github.com/riportdev/riport/share/query"
)

func TestMultiJobsSqliteProvider(t *testing.T) {
	ctx := context.Background()
	jobsDB, err := sqlite.New(":memory:", jobs.AssetNames(), jobs.Asset, DataSourceOptions)
	require.NoError(t, err)
	p := NewSqliteProvider(jobsDB, testLog)
	defer p.Close()

	// verify job summaries not found
	gotJSs, err := p.GetMultiJobSummaries(ctx, &query.ListOptions{})
	require.NoError(t, err)
	require.Empty(t, gotJSs)

	// add jobs
	t1 := time.Now().UTC()
	job1 := jb.NewMulti(t).JID("1111").StartedAt(t1.Add(-time.Hour)).WithJobs().WithSudo().Build()
	job2 := jb.NewMulti(t).JID("2222").StartedAt(t1).WithCwd("/root").Build() // jid used to check the order by
	job3 := jb.NewMulti(t).JID("3333").StartedAt(t1).Build()
	require.NoError(t, p.SaveMultiJob(job1))
	for _, j := range job1.Jobs {
		require.NoError(t, p.SaveJob(j))
	}
	require.NoError(t, p.SaveMultiJob(job2))
	require.NoError(t, p.SaveMultiJob(job3))

	// verify added jobs
	gotJob1, err := p.GetMultiJob(ctx, job1.JID)
	require.NoError(t, err)
	require.NotNil(t, gotJob1)
	assert.EqualValues(t, job1, gotJob1)

	gotJob2, err := p.GetMultiJob(ctx, job2.JID)
	require.NoError(t, err)
	require.NotNil(t, gotJob2)
	assert.Equal(t, job2, gotJob2)

	gotJob3, err := p.GetMultiJob(ctx, job3.JID)
	require.NoError(t, err)
	require.NotNil(t, gotJob3)
	assert.Equal(t, job3, gotJob3)

	// verify child jobs
	childJobs, err := p.List(ctx, &query.ListOptions{Filters: []query.FilterOption{{Column: []string{"multi_job_id"}, Values: []string{job1.JID}}}})
	require.NoError(t, err)
	assert.ElementsMatch(t, job1.Jobs, childJobs)

	// verify not found job
	gotJob4, err := p.GetMultiJob(ctx, "unknown-jid")
	require.NoError(t, err)
	require.Nil(t, gotJob4)

	// verify job summaries
	gotJSs, err = p.GetMultiJobSummaries(ctx, &query.ListOptions{})
	require.NoError(t, err)
	assert.EqualValues(t, []*models.MultiJobSummary{&job2.MultiJobSummary, &job3.MultiJobSummary, &job1.MultiJobSummary}, gotJSs)

	// verify jobs count
	count, err := p.CountMultiJobs(ctx, &query.ListOptions{})
	require.NoError(t, err)
	assert.Equal(t, 3, count)

	// verify job update
	job1.Interpreter = "cmd"
	job1.Concurrent = true
	job1.StartedAt = t1.Add(time.Second)

	require.NoError(t, p.SaveMultiJob(job1))
	gotJob1, err = p.GetMultiJob(ctx, job1.JID)
	require.NoError(t, err)
	require.NotNil(t, gotJob1)
	assert.Equal(t, job1, gotJob1)

	gotJSs, err = p.GetMultiJobSummaries(ctx, &query.ListOptions{})
	require.NoError(t, err)
	assert.EqualValues(t, []*models.MultiJobSummary{&job1.MultiJobSummary, &job2.MultiJobSummary, &job3.MultiJobSummary}, gotJSs)
}
