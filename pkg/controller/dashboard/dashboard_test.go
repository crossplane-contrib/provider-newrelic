/*
Copyright 2020 The Crossplane Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

   http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package dashboard

import (
	"testing"

	"github.com/crossplane/crossplane-runtime/pkg/meta"
	"github.com/google/go-cmp/cmp"
	"github.com/newrelic/newrelic-client-go/v2/pkg/dashboards"
	"github.com/newrelic/newrelic-client-go/v2/pkg/entities"
	"github.com/openlyinc/pointy"

	"github.com/crossplane-contrib/provider-newrelic/apis/dashboard/v1alpha1"
)

type DashboardModifier func(dashboard *v1alpha1.Dashboard)

func withVariables(m []v1alpha1.DashboardVariable) DashboardModifier {
	return func(d *v1alpha1.Dashboard) {
		d.Spec.ForProvider.Variables = m
	}
}

func withPages(m []v1alpha1.DashboardPage) DashboardModifier {
	return func(d *v1alpha1.Dashboard) {
		d.Spec.ForProvider.Pages = m
	}
}

func Dashboard(m ...DashboardModifier) *v1alpha1.Dashboard {
	cr := &v1alpha1.Dashboard{
		Spec: v1alpha1.DashboardSpec{
			ForProvider: v1alpha1.DashboardParameters{
				AccountID: 1,
				Name:      "test_dashboard",
				Variables: []v1alpha1.DashboardVariable{},
				Pages: []v1alpha1.DashboardPage{
					{Name: "test_dashboard_page_1",
						GUID: "PAGE1GUID",
						Widgets: []v1alpha1.DashboardWidget{
							{ID: pointy.String("test_dashboard_widget_1"),
								Title: "dashboard_title_1", Layout: v1alpha1.DashboardWidgetLayout{
									Column: 1,
									Height: 1,
									Row:    1,
									Width:  1,
								},
								Visualization: v1alpha1.DashboardWidgetVisualization{ID: "viz.area"},
								RawConfiguration: &v1alpha1.DashboardWidgetRawConfiguration{
									NRQLQueries: &[]v1alpha1.DashboardWidgetNRQLQueryInput{{AccountID: "1", Query: "Select * FROM Metric"}},
								},
							},
							{ID: pointy.String("test_dashboard_widget_2"),
								Title: "dashboard_title_2", Layout: v1alpha1.DashboardWidgetLayout{
									Column: 1,
									Height: 1,
									Row:    1,
									Width:  1,
								},
								Visualization: v1alpha1.DashboardWidgetVisualization{ID: "viz.area"},
								RawConfiguration: &v1alpha1.DashboardWidgetRawConfiguration{
									NRQLQueries: &[]v1alpha1.DashboardWidgetNRQLQueryInput{{AccountID: "1", Query: "Select * FROM Metric"}},
								},
							},
						},
					},
					{Name: "test_dashboard_page_2",
						GUID: "PAGE1GUID",
						Widgets: []v1alpha1.DashboardWidget{
							{ID: pointy.String("test_dashboard_widget_1"),
								Title: "dashboard_title_1", Layout: v1alpha1.DashboardWidgetLayout{
									Column: 1,
									Height: 1,
									Row:    1,
									Width:  1,
								},
								Visualization: v1alpha1.DashboardWidgetVisualization{ID: "viz.area"},
								RawConfiguration: &v1alpha1.DashboardWidgetRawConfiguration{
									NRQLQueries: &[]v1alpha1.DashboardWidgetNRQLQueryInput{{AccountID: "1", Query: "Select * FROM Metric"}},
								},
							},
						},
					},
				},
			},
		},
	}
	meta.SetExternalName(cr, "1")
	for _, f := range m {
		f(cr)
	}
	cr.Spec.ForProvider.GUID = "1375108"
	return cr
}

func DashboardBillboard(m ...DashboardModifier) *v1alpha1.Dashboard {
	cr := &v1alpha1.Dashboard{
		Spec: v1alpha1.DashboardSpec{
			ForProvider: v1alpha1.DashboardParameters{
				AccountID: 1,
				Name:      "test_dashboard",
				Variables: []v1alpha1.DashboardVariable{},
				Pages: []v1alpha1.DashboardPage{
					{Name: "test_dashboard_page_1",
						GUID: "PAGE1GUID",
						Widgets: []v1alpha1.DashboardWidget{
							{ID: pointy.String("test_dashboard_widget_1"),
								Title: "dashboard_title_1", Layout: v1alpha1.DashboardWidgetLayout{
									Column: 1,
									Height: 1,
									Row:    1,
									Width:  1,
								},
								Visualization: v1alpha1.DashboardWidgetVisualization{ID: "viz.area"},
								RawConfiguration: &v1alpha1.DashboardWidgetRawConfiguration{
									NRQLQueries: &[]v1alpha1.DashboardWidgetNRQLQueryInput{{AccountID: "1", Query: "Select * FROM Metric"}},
									Thresholds: []v1alpha1.DashboardBillboardWidgetThresholdInput{{AlertSeverity: pointy.String("Warning"), Value: pointy.Float64(50)},
										{AlertSeverity: pointy.String("Critical"), Value: pointy.Float64(90)},
									},
								},
							},
						},
					},
				},
			},
		},
	}
	meta.SetExternalName(cr, "1")
	for _, f := range m {
		f(cr)
	}
	cr.Spec.ForProvider.GUID = "1375108"
	return cr
}

func TestIsUpToDate(t *testing.T) {

	type args struct {
		cr v1alpha1.Dashboard
		nr entities.DashboardEntity
	}

	type want struct {
		expected bool
	}

	cases := map[string]struct {
		args args
		want want
	}{
		"DiffNameFalse": {
			args: args{cr: *Dashboard(),
				nr: entities.DashboardEntity{
					Name:        "test_dashboard_diff_name",
					Description: "",
					Pages: []entities.DashboardPage{{
						Description: "",
						GUID:        "PAGE1GUID",
						Name:        "test_dashboard_page_1",
						Widgets: []entities.DashboardWidget{
							{ID: "test_dashboard_widget_1",
								Title: "dashboard_title_1", Layout: entities.DashboardWidgetLayout{
									Column: 1,
									Height: 1,
									Row:    1,
									Width:  1,
								},
								Visualization: entities.DashboardWidgetVisualization{ID: "viz.area"},
								Configuration: entities.DashboardWidgetConfiguration{
									Area: entities.DashboardAreaWidgetConfiguration{NRQLQueries: []entities.DashboardWidgetNRQLQuery{{AccountID: 1, Query: "Select * FROM Metric"}}},
								},
							},
							{ID: "test_dashboard_widget_2",
								Title: "dashboard_title_2", Layout: entities.DashboardWidgetLayout{
									Column: 1,
									Height: 1,
									Row:    1,
									Width:  1,
								},
								Visualization: entities.DashboardWidgetVisualization{ID: "viz.area"},
								Configuration: entities.DashboardWidgetConfiguration{
									Area: entities.DashboardAreaWidgetConfiguration{NRQLQueries: []entities.DashboardWidgetNRQLQuery{{AccountID: 1, Query: "Select * FROM Metric"}}},
								},
							},
						},
					},
						{
							Description: "",
							GUID:        "PAGE1GUID",
							Name:        "test_dashboard_page_2",
							Widgets: []entities.DashboardWidget{
								{ID: "test_dashboard_widget_1",
									Title: "dashboard_title_1", Layout: entities.DashboardWidgetLayout{
										Column: 1,
										Height: 1,
										Row:    1,
										Width:  1,
									},
									Visualization: entities.DashboardWidgetVisualization{ID: "viz.area"},
									Configuration: entities.DashboardWidgetConfiguration{
										Area: entities.DashboardAreaWidgetConfiguration{NRQLQueries: []entities.DashboardWidgetNRQLQuery{{AccountID: 1, Query: "Select * FROM Metric"}}},
									},
								},
							},
						}},
					Permissions: "PUBLIC_READ_WRITE",
				},
			},
			want: want{expected: false},
		},
		"VariablesSame": {
			args: args{cr: *Dashboard(withPages([]v1alpha1.DashboardPage{}),
				withVariables([]v1alpha1.DashboardVariable{{
					DefaultValues:    &[]v1alpha1.DashboardVariableDefaultItem{{Value: v1alpha1.DashboardVariableDefaultValue{String: "*"}}},
					IsMultiSelection: true,
					NRQLQuery: &v1alpha1.DashboardVariableNRQLQuery{AccountIDs: []int{1},
						Query: "SELECT count(*) from Metric"},
					Name:                "TestVariable",
					ReplacementStrategy: "STRING",
					Title:               "TestTitle",
					Type:                "NRQL",
				},
				})),
				nr: entities.DashboardEntity{
					Name:        "test_dashboard",
					Description: "",
					Pages:       []entities.DashboardPage{},
					Permissions: "PUBLIC_READ_WRITE",
					Variables: []entities.DashboardVariable{{
						DefaultValues:    &[]entities.DashboardVariableDefaultItem{{Value: entities.DashboardVariableDefaultValue{String: "*"}}},
						IsMultiSelection: true,
						NRQLQuery: &entities.DashboardVariableNRQLQuery{AccountIDs: []int{1},
							Query: "SELECT count(*) from Metric"},
						Name:                "TestVariable",
						ReplacementStrategy: "STRING",
						Title:               "TestTitle",
						Type:                "NRQL",
					},
					},
				}},
			want: want{expected: true},
		},
		"DiffVariablesFalse": {
			args: args{cr: *Dashboard(withPages([]v1alpha1.DashboardPage{}),
				withVariables([]v1alpha1.DashboardVariable{{
					DefaultValues:    &[]v1alpha1.DashboardVariableDefaultItem{{Value: v1alpha1.DashboardVariableDefaultValue{String: "*"}}},
					IsMultiSelection: true,
					NRQLQuery: &v1alpha1.DashboardVariableNRQLQuery{AccountIDs: []int{1},
						Query: "SELECT count(*) from Metric"},
					Name:                "TestVariable",
					ReplacementStrategy: "STRING",
					Title:               "TestTitle",
					Type:                "NRQL",
				},
				})),
				nr: entities.DashboardEntity{
					Name:        "test_dashboard",
					Description: "",
					Pages:       []entities.DashboardPage{},
					Permissions: "PUBLIC_READ_WRITE",
					Variables: []entities.DashboardVariable{{
						DefaultValues:    &[]entities.DashboardVariableDefaultItem{{Value: entities.DashboardVariableDefaultValue{String: "*"}}},
						IsMultiSelection: true,
						NRQLQuery: &entities.DashboardVariableNRQLQuery{AccountIDs: []int{1},
							Query: "SELECT count(*) from Metric WHERE environment='different'"},
						Name:                "TestVariable",
						ReplacementStrategy: "STRING",
						Title:               "TestTitle",
						Type:                "NRQL",
					},
					},
				}},
			want: want{expected: false},
		},
		"BillboardOutOfOrderThresholdTrue": {
			args: args{cr: *DashboardBillboard(),
				nr: entities.DashboardEntity{
					Name:        "test_dashboard",
					Description: "",
					Pages: []entities.DashboardPage{{
						Description: "",
						GUID:        "PAGE1GUID",
						Name:        "test_dashboard_page_1",
						Widgets: []entities.DashboardWidget{
							{ID: "test_dashboard_widget_1",
								Title: "dashboard_title_1", Layout: entities.DashboardWidgetLayout{
									Column: 1,
									Height: 1,
									Row:    1,
									Width:  1,
								},
								Visualization: entities.DashboardWidgetVisualization{ID: "viz.area"},
								Configuration: entities.DashboardWidgetConfiguration{
									Billboard: entities.DashboardBillboardWidgetConfiguration{
										NRQLQueries: []entities.DashboardWidgetNRQLQuery{{AccountID: 1, Query: "Select * FROM Metric"}},
										Thresholds: []entities.DashboardBillboardWidgetThreshold{
											// These are out of order
											{AlertSeverity: "Critical", Value: 90},
											{AlertSeverity: "Warning", Value: 50},
										},
									},
								},
							},
						},
					}},
					Permissions: "PUBLIC_READ_WRITE",
				},
			},
			want: want{expected: true},
		},
		"RawConfigurationIgnoreTrue": {
			args: args{cr: *Dashboard(),
				nr: entities.DashboardEntity{
					Name:        "test_dashboard",
					Description: "",
					Pages: []entities.DashboardPage{{
						Description: "",
						GUID:        "PAGE1GUID",
						Name:        "test_dashboard_page_1",
						Widgets: []entities.DashboardWidget{
							{ID: "test_dashboard_widget_1",
								Title: "dashboard_title_1", Layout: entities.DashboardWidgetLayout{
									Column: 1,
									Height: 1,
									Row:    1,
									Width:  1,
								},
								Visualization: entities.DashboardWidgetVisualization{ID: "viz.area"},
								Configuration: entities.DashboardWidgetConfiguration{
									Area: entities.DashboardAreaWidgetConfiguration{NRQLQueries: []entities.DashboardWidgetNRQLQuery{{AccountID: 1, Query: "Select * FROM Metric"}}},
								},
							},
							{ID: "test_dashboard_widget_2",
								Title: "dashboard_title_2", Layout: entities.DashboardWidgetLayout{
									Column: 1,
									Height: 1,
									Row:    1,
									Width:  1,
								},
								Visualization: entities.DashboardWidgetVisualization{ID: "viz.area"},
								Configuration: entities.DashboardWidgetConfiguration{
									Area: entities.DashboardAreaWidgetConfiguration{NRQLQueries: []entities.DashboardWidgetNRQLQuery{{AccountID: 1, Query: "Select * FROM Metric"}}},
								},
							},
						},
					},
						{
							Description: "",
							GUID:        "PAGE1GUID",
							Name:        "test_dashboard_page_2",
							Widgets: []entities.DashboardWidget{
								{ID: "test_dashboard_widget_1",
									Title: "dashboard_title_1", Layout: entities.DashboardWidgetLayout{
										Column: 1,
										Height: 1,
										Row:    1,
										Width:  1,
									},
									Visualization:    entities.DashboardWidgetVisualization{ID: "viz.area"},
									RawConfiguration: entities.DashboardWidgetRawConfiguration("{\n  \"facet\": {\n    \"showOtherSeries\": false\n    },\n  \"legend\": {\n    \"enabled\": true\n    },\"nrqlQueries\": [\n                    {\n                      \"accountId\": 1448011,\n                      \"query\": \"SELECT count(aws.states.ExecutionsStarted) FROM Metric WHERE tags.AppName = ''CustomerPipeline'' AND tags.EnvName = ''prod-us'' TIMESERIES FACET tags.tf_ignore_pipeline_version \"\n    }\n  ]\n}"),
									Configuration: entities.DashboardWidgetConfiguration{
										Area: entities.DashboardAreaWidgetConfiguration{NRQLQueries: []entities.DashboardWidgetNRQLQuery{{AccountID: 1, Query: "Select * FROM Metric"}}},
									},
								},
							},
						}},
					Permissions: "PUBLIC_READ_WRITE",
				},
			},
			want: want{expected: true},
		},
		"SameOutOfOrderTrue": {
			args: args{cr: *Dashboard(),
				nr: entities.DashboardEntity{
					Name:        "test_dashboard",
					Description: "",
					Pages: []entities.DashboardPage{{
						Description: "",
						GUID:        "PAGE1GUID",
						Name:        "test_dashboard_page_1",
						Widgets: []entities.DashboardWidget{
							{ID: "test_dashboard_widget_2",
								Title: "dashboard_title_2", Layout: entities.DashboardWidgetLayout{
									Column: 1,
									Height: 1,
									Row:    1,
									Width:  1,
								},
								Visualization: entities.DashboardWidgetVisualization{ID: "viz.area"},
								Configuration: entities.DashboardWidgetConfiguration{
									Area: entities.DashboardAreaWidgetConfiguration{NRQLQueries: []entities.DashboardWidgetNRQLQuery{{AccountID: 1, Query: "Select * FROM Metric"}}},
								},
							},
							{ID: "test_dashboard_widget_1",
								Title: "dashboard_title_1", Layout: entities.DashboardWidgetLayout{
									Column: 1,
									Height: 1,
									Row:    1,
									Width:  1,
								},
								Visualization: entities.DashboardWidgetVisualization{ID: "viz.area"},
								Configuration: entities.DashboardWidgetConfiguration{
									Area: entities.DashboardAreaWidgetConfiguration{NRQLQueries: []entities.DashboardWidgetNRQLQuery{{AccountID: 1, Query: "Select * FROM Metric"}}},
								},
							},
						},
					},
						{
							Description: "",
							GUID:        "PAGE1GUID",
							Name:        "test_dashboard_page_2",
							Widgets: []entities.DashboardWidget{
								{ID: "test_dashboard_widget_1",
									Title: "dashboard_title_1", Layout: entities.DashboardWidgetLayout{
										Column: 1,
										Height: 1,
										Row:    1,
										Width:  1,
									},
									Visualization: entities.DashboardWidgetVisualization{ID: "viz.area"},
									Configuration: entities.DashboardWidgetConfiguration{
										Area: entities.DashboardAreaWidgetConfiguration{NRQLQueries: []entities.DashboardWidgetNRQLQuery{{AccountID: 1, Query: "Select * FROM Metric"}}},
									},
								},
							},
						}},
					Permissions: "PUBLIC_READ_WRITE",
				},
			},
			want: want{expected: true},
		},
		"SameTrue": {
			args: args{cr: *Dashboard(),
				nr: entities.DashboardEntity{
					Name:        "test_dashboard",
					Description: "",
					Pages: []entities.DashboardPage{{
						Description: "",
						GUID:        "PAGE1GUID",
						Name:        "test_dashboard_page_1",
						Widgets: []entities.DashboardWidget{
							{ID: "test_dashboard_widget_1",
								Title: "dashboard_title_1", Layout: entities.DashboardWidgetLayout{
									Column: 1,
									Height: 1,
									Row:    1,
									Width:  1,
								},
								Visualization: entities.DashboardWidgetVisualization{ID: "viz.area"},
								Configuration: entities.DashboardWidgetConfiguration{
									Area: entities.DashboardAreaWidgetConfiguration{NRQLQueries: []entities.DashboardWidgetNRQLQuery{{AccountID: 1, Query: "Select * FROM Metric"}}},
								},
							},
							{ID: "test_dashboard_widget_2",
								Title: "dashboard_title_2", Layout: entities.DashboardWidgetLayout{
									Column: 1,
									Height: 1,
									Row:    1,
									Width:  1,
								},
								Visualization: entities.DashboardWidgetVisualization{ID: "viz.area"},
								Configuration: entities.DashboardWidgetConfiguration{
									Area: entities.DashboardAreaWidgetConfiguration{NRQLQueries: []entities.DashboardWidgetNRQLQuery{{AccountID: 1, Query: "Select * FROM Metric"}}},
								},
							},
						},
					},
						{
							Description: "",
							GUID:        "PAGE1GUID",
							Name:        "test_dashboard_page_2",
							Widgets: []entities.DashboardWidget{
								{ID: "test_dashboard_widget_1",
									Title: "dashboard_title_1", Layout: entities.DashboardWidgetLayout{
										Column: 1,
										Height: 1,
										Row:    1,
										Width:  1,
									},
									Visualization: entities.DashboardWidgetVisualization{ID: "viz.area"},
									Configuration: entities.DashboardWidgetConfiguration{
										Area: entities.DashboardAreaWidgetConfiguration{NRQLQueries: []entities.DashboardWidgetNRQLQuery{{AccountID: 1, Query: "Select * FROM Metric"}}},
									},
								},
							},
						}},
					Permissions: "PUBLIC_READ_WRITE",
				},
			},
			want: want{expected: true},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			got := IsUpToDate(&tc.args.cr, tc.args.nr)
			if diff := cmp.Diff(tc.want.expected, got); diff != "" {
				t.Errorf("e.TestIsUpToDate(...): -want, +got:\n%s\n", diff)
			}
		})
	}
}

func TestGenerateDashboardPageInputFromEntity(t *testing.T) {

	type args struct {
		nr entities.DashboardEntity
	}

	type want struct {
		expected dashboards.DashboardInput
	}

	cases := map[string]struct {
		args args
		want want
	}{
		"SameOutOfOrderTrue": {
			args: args{nr: entities.DashboardEntity{
				Name:        "test_dashboard",
				Description: "",
				Pages: []entities.DashboardPage{{
					Description: "",
					GUID:        "PAGE1GUID",
					Name:        "test_dashboard_page_1",
					Widgets: []entities.DashboardWidget{
						{ID: "test_dashboard_widget_2",
							Title: "dashboard_title_2", Layout: entities.DashboardWidgetLayout{
								Column: 1,
								Height: 1,
								Row:    1,
								Width:  1,
							},
							Visualization:    entities.DashboardWidgetVisualization{ID: "viz.area"},
							RawConfiguration: []byte("\"\\\"nrqlQueries\\\": [\\n{\\n\\\"accountId\\\": 1,\\n \\\"query\\\": \\\"\\\"Select * FROM Metric \\\"\\n}\\n]\""),
						},
						{ID: "test_dashboard_widget_1",
							Title: "dashboard_title_1", Layout: entities.DashboardWidgetLayout{
								Column: 1,
								Height: 1,
								Row:    1,
								Width:  1,
							},
							Visualization:    entities.DashboardWidgetVisualization{ID: "viz.area"},
							RawConfiguration: []byte("\"\\\"nrqlQueries\\\": [\\n{\\n\\\"accountId\\\": 1,\\n \\\"query\\\": \\\"\\\"Select * FROM Metric \\\"\\n}\\n]\""),
						},
					},
				},
					{
						Description: "",
						GUID:        "PAGE1GUID",
						Name:        "test_dashboard_page_2",
						Widgets: []entities.DashboardWidget{
							{ID: "test_dashboard_widget_1",
								Title: "dashboard_title_1", Layout: entities.DashboardWidgetLayout{
									Column: 1,
									Height: 1,
									Row:    1,
									Width:  1,
								},
								Visualization:    entities.DashboardWidgetVisualization{ID: "viz.area"},
								RawConfiguration: []byte("\"\\\"nrqlQueries\\\": [\\n{\\n\\\"accountId\\\": 1,\\n \\\"query\\\": \\\"\\\"Select * FROM Metric \\\"\\n}\\n]\""),
							},
						},
					}},
				Permissions: "PUBLIC_READ_WRITE",
			},
			},
			want: want{expected: dashboards.DashboardInput{
				Name:        "test_dashboard",
				Description: "",
				Pages: []dashboards.DashboardPageInput{{
					Description: "",
					GUID:        "PAGE1GUID",
					Name:        "test_dashboard_page_1",
					Widgets: []dashboards.DashboardWidgetInput{
						{ID: "test_dashboard_widget_1",
							Title: "dashboard_title_1", Layout: dashboards.DashboardWidgetLayoutInput{
								Column: 1,
								Height: 1,
								Row:    1,
								Width:  1,
							},
							Visualization:    dashboards.DashboardWidgetVisualizationInput{ID: "viz.area"},
							RawConfiguration: []byte("\"\\\"nrqlQueries\\\": [\\n{\\n\\\"accountId\\\": 1,\\n \\\"query\\\": \\\"\\\"Select * FROM Metric \\\"\\n}\\n]\""),
						},
						{ID: "test_dashboard_widget_2",
							Title: "dashboard_title_2", Layout: dashboards.DashboardWidgetLayoutInput{
								Column: 1,
								Height: 1,
								Row:    1,
								Width:  1,
							},
							Visualization:    dashboards.DashboardWidgetVisualizationInput{ID: "viz.area"},
							RawConfiguration: []byte("\"\\\"nrqlQueries\\\": [\\n{\\n\\\"accountId\\\": 1,\\n \\\"query\\\": \\\"\\\"Select * FROM Metric \\\"\\n}\\n]\""),
						},
					},
				},
					{
						Description: "",
						GUID:        "PAGE1GUID",
						Name:        "test_dashboard_page_2",
						Widgets: []dashboards.DashboardWidgetInput{
							{ID: "test_dashboard_widget_1",
								Title: "dashboard_title_1", Layout: dashboards.DashboardWidgetLayoutInput{
									Column: 1,
									Height: 1,
									Row:    1,
									Width:  1,
								},
								Visualization:    dashboards.DashboardWidgetVisualizationInput{ID: "viz.area"},
								RawConfiguration: []byte("\"\\\"nrqlQueries\\\": [\\n{\\n\\\"accountId\\\": 1,\\n \\\"query\\\": \\\"\\\"Select * FROM Metric \\\"\\n}\\n]\""),
							},
						},
					}},
				Permissions: "PUBLIC_READ_WRITE",
				Variables:   []dashboards.DashboardVariableInput{},
			},
			}},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			got := GenerateDashboardInputFromEntity(tc.args.nr)
			if diff := cmp.Diff(tc.want.expected, got); diff != "" {
				t.Errorf("e.GenerateDashboardPageInput(...): -want, +got:\n%s\n", diff)
			}
		})
	}
}
