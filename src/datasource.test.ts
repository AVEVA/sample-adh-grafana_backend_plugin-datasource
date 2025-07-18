import { DataQueryRequest, DataSourceInstanceSettings, FieldType, MutableDataFrame } from '@grafana/data';
import { SdsDataSourceOptions, SdsDataSourceType, SdsQuery } from 'types';
import { DataSource } from 'datasource';
import { Observable } from 'rxjs';

jest.mock('@grafana/runtime', () => {
  const original = jest.requireActual('@grafana/runtime');
  return {
    ...original,
    getTemplateSrv: () => ({
      getVariables: () => [],
      replace: (s: string) => s,
    }),
  };
});

describe('DataSource', () => {
  const url = 'URL';
  const edsPort = 'PORT';
  const resource = 'URL';
  const apiVersion = 'VERSION';
  const tenantId = 'TENANT';
  const clientId = 'CLIENT';
  const oauthPassThru = false;
  const namespaceId = 'NAMESPACE';
  const useCommunity = false;
  const communityId = 'COMMUNITY';
  const adhSettings: DataSourceInstanceSettings<SdsDataSourceOptions> = {
    id: 0,
    uid: '',
    name: '',
    access: 'direct',
    type: '',
    url,
    meta: null as any,
    jsonData: {
      type: SdsDataSourceType.ADH,
      edsPort: edsPort,
      resource: resource,
      apiVersion: apiVersion,
      tenantId: tenantId,
      clientId: clientId,
      oauthPassThru: oauthPassThru,
      namespaceId: namespaceId,
      useCommunity: useCommunity,
      communityId: communityId,
    },
    readOnly: false,
  };

  describe('constructor', () => {
    it('should use passed in data source information', () => {
      const datasource = new DataSource(adhSettings);
      expect(datasource.type).toEqual(SdsDataSourceType.ADH);
      expect(datasource.edsPort).toEqual(edsPort);
    });
  });

  describe('queryEDS', () => {
    it('should query with the expected parameters', (done) => {
      const options = {
        range: {
          from: {
            utc: () => ({
              format: () => 'FROM',
            }),
          },
          to: {
            utc: () => ({
              format: () => 'TO',
            }),
          },
        },
        targets: [
          {
            refId: 'REFID',
            name: 'STREAM',
            querytext: 'QUERYTEXT',
            collection: 'COLLECTION',
            id: 'ID',
          },
        ],
      } as unknown as DataQueryRequest<SdsQuery>;

      const datasource = new DataSource(adhSettings);

      const results = datasource.queryEDS(options);

      results.subscribe({
        next(result) {
          expect(JSON.stringify(result)).toEqual(
            JSON.stringify({
              data: [
                new MutableDataFrame({
                  refId: 'REFID',
                  name: 'STREAM',
                  fields: [
                    {
                      name: 'TimeStamp',
                      type: FieldType.time,
                      values: [Date.parse('2020-01-01')],
                    },
                    {
                      name: 'Boolean',
                      type: FieldType.number,
                      values: [1],
                    },
                    {
                      name: 'Number',
                      type: FieldType.number,
                      values: [1],
                    },
                    {
                      name: 'String',
                      type: FieldType.string,
                      values: ['A'],
                    },
                  ],
                }),
              ],
            })
          );
          done();
        },
      });
    });
  });

  describe('getStreams', () => {
    it('should query for streams', (done) => {
      const datasource = new DataSource(adhSettings);

      datasource.query = jest.fn(() => {
        return new Observable((subscriber) => {
          subscriber.next({
            data: [
              new MutableDataFrame({
                refId: 'REFID',
                name: 'STREAM',
                fields: [
                  {
                    name: 'Id',
                    type: FieldType.string,
                    values: ['Id1', 'Id2'],
                  },
                  {
                    name: 'Name',
                    type: FieldType.string,
                    values: ['Name1', 'Name2'],
                  },
                ],
              }),
            ],
          });

          subscriber.complete();
        });
      });

      const results = datasource.getStreams('QUERY', () => {});

      results.then((r) => {
        expect(r).toEqual([
          { value: 'Id1', label: 'Name1' },
          { value: 'Id2', label: 'Name2' },
        ]);
        done();
      });
    });
  });
});
