import { DataSourcePlugin } from '@grafana/data';
import { DataSource } from './datasource';
import { ConfigEditor } from './components/ConfigEditor';
import { QueryEditor } from './components/QueryEditor';
import { SdsQuery, SdsDataSourceOptions, SdsDataSourceSecureOptions } from './types';

export const plugin = new DataSourcePlugin<DataSource, SdsQuery, SdsDataSourceOptions, SdsDataSourceSecureOptions>(DataSource)
  .setConfigEditor(ConfigEditor)
  .setQueryEditor(QueryEditor);
