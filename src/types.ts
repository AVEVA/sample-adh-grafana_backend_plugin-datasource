import { DataSourceJsonData } from '@grafana/data';
import { DataQuery } from '@grafana/schema';

export enum SdsDataSourceType {
  CDS = 'ADH',
  EDS = 'EDS',
}

export interface SdsQuery extends DataQuery {
  collection: string;
  queryText: string;
  id: string;
  name: string;
}

export const defaultQuery: Partial<SdsQuery> = {
  collection: 'streams',
  queryText: '',
  id: '',
  name: '',
};

export interface SdsDataSourceOptions extends DataSourceJsonData {
  type: SdsDataSourceType;
  edsPort: string;
  resource: string;
  apiVersion: string;
  accountId: string;
  clientId: string;
  oauthPassThru: boolean;
  sdsId: string;
}

export interface SdsDataSourceSecureOptions {
  clientSecret: string;
}

