import { DataSourceInstanceSettings, DataQueryRequest, DataQueryResponse, FieldType, MutableDataFrame, DataFrame, SelectableValue } from '@grafana/data';
import { DataSourceWithBackend, FetchResponse, getBackendSrv } from '@grafana/runtime';

import { defaultQuery, SdsDataSourceOptions, SdsDataSourceType, SdsQuery } from './types';
import { Observable, zip, map, lastValueFrom } from 'rxjs';
import { Dispatch, SetStateAction } from 'react';

export class DataSource extends DataSourceWithBackend<SdsQuery, SdsDataSourceOptions> {
  type: SdsDataSourceType;
  edsPort: string;
  constructor(instanceSettings: DataSourceInstanceSettings<SdsDataSourceOptions>) {
    super(instanceSettings);
    this.type = instanceSettings.jsonData?.type || SdsDataSourceType.ADH;
    this.edsPort = instanceSettings.jsonData?.edsPort || '5590';
  }

  queryEDS(request: DataQueryRequest<SdsQuery>): Observable<DataQueryResponse> {
    const from = request.range.from.utc().format();
    const to = request.range.to.utc().format();

    const requests = request.targets.map((target) => {
      if (target.id === '') {
        return getBackendSrv().fetch({
          url: `http://localhost:${this.edsPort}/api/v1/tenants/default/namespaces/default/streams?query=${target.queryText}`,
          method: 'GET',
        });
      } else {
        return getBackendSrv().fetch({
          url: `http://localhost:${this.edsPort}/api/v1/tenants/default/namespaces/default/streams/${target.id}/data?startIndex=${from}&endIndex=${to}`,
          method: 'GET',
        });
      }
    });

    return zip(requests).pipe(
      map((responses) => {
        let i = 0;
        const data = responses.map((r: FetchResponse) => {
          if (!r || !r.data.length) {
            return new MutableDataFrame();
          }

          const target = request.targets[i];
          i++;
          return new MutableDataFrame({
            refId: target.refId,
            name: target.name,
            fields: Object.keys(r.data[0]).map((name) => {
              const val0 = r.data[0][name];
              const date = Date.parse(val0);
              const num = Number(val0);
              const type =
                typeof val0 === 'string' && !isNaN(date)
                  ? FieldType.time
                  : val0 === true || val0 === false
                  ? FieldType.boolean
                  : !isNaN(num)
                  ? FieldType.number
                  : FieldType.string;

              let values = [];
              if (type === FieldType.boolean) {
                values = r.data.map((d: any) => {
                  return d[name]?.toString().toLowerCase() === 'true' ? 1 : 0;
                });
              } else {
                values = r.data.map((d: any) => (type === FieldType.time ? Date.parse(d[name]) : d[name]));
              }

              return {
                name,
                values: values,
                type: type === FieldType.boolean ? FieldType.number : type,
              };
            }),
          });
        });

        return { data };
      })
    );
  }

query(request: DataQueryRequest<SdsQuery>): Observable<DataQueryResponse> {
    if (this.type === SdsDataSourceType.ADH) {
      return super.query(request);
    } else {
      return this.queryEDS(request);
    }
  }

  async getStreams(
    query: string,
    stateAction: Dispatch<SetStateAction<boolean | Array<SelectableValue<string>>>>
  ): Promise<Array<SelectableValue<string>>> {
    const observableResponse = this.query({
      targets: [{ ...defaultQuery, refId: 'sds-stream-autocomplete', queryText: query, collection: 'streams', id: '' }],
    } as DataQueryRequest<SdsQuery>);

    const response = await lastValueFrom(observableResponse);

    if (!Array.isArray(response?.data) || response?.data.length === 0) {
      return [];
    }

    const dataFrame = response.data[0] as DataFrame;

    if (!Array.isArray(dataFrame?.fields) || dataFrame?.fields.length === 0) {
      return [];
    }

    const ids = dataFrame.fields[dataFrame.fields.findIndex((field) => field.name === 'Id')].values.toArray();
    const names = dataFrame.fields[dataFrame.fields.findIndex((field) => field.name === 'Name')].values.toArray();

    const selectables = [];
    for (let i = 0; i < ids.length; i++) {
      selectables.push({ value: ids[i], label: names[i] });
    }

    // Set state to persist selectables list
    stateAction(selectables);

    return selectables;
  }
}
