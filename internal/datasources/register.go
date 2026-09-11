package datasources

import (
	"github.com/hashicorp/terraform-plugin-framework/datasource"
)

// AllDataSources returns all data source constructors provided by this package.
func AllDataSources() []func() datasource.DataSource {
	return []func() datasource.DataSource{
		NewConfigStatusDataSource,
		NewConfigPreviewDataSource,
		NewConfigDiffDataSource,
		NewConfigSnapshotsDataSource,
		NewBackendDataSource,
		NewListenerDataSource,
		NewSystemStatsDataSource,
		NewHaproxyStatsDataSource,
		NewHealthDataSource,
		NewAuditEventsDataSource,
		NewRecentLogsDataSource,
		NewStickTablesDataSource,
		NewStickTableDataSource,
		NewValkeyInfoDataSource,
		NewValkeyNamespacesDataSource,
		NewGeoipStatusDataSource,
		NewAsnLookupDataSource,
		NewSslLabsScansDataSource,
		NewMcpGatewayStatusDataSource,
		NewMcpConfigStatusDataSource,
		NewMcpEventsDataSource,
		NewMcpMarketplaceSearchDataSource,
	}
}
