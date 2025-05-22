import React, { SyntheticEvent } from 'react';
import {
  DataSourcePluginOptionsEditorProps,
  onUpdateDatasourceSecureJsonDataOption,
  onUpdateDatasourceJsonDataOption,
  onUpdateDatasourceJsonDataOptionChecked,
  onUpdateDatasourceJsonDataOptionSelect,
} from '@grafana/data';
import { Select, InlineSwitch, InlineField, Input, InlineFieldRow } from '@grafana/ui';
import { SdsDataSourceOptions, SdsDataSourceType, SdsDataSourceSecureOptions } from '../types';

interface Props extends DataSourcePluginOptionsEditorProps<SdsDataSourceOptions, SdsDataSourceSecureOptions> {}

export const ConfigEditor = (props: Props) => {
  const typeLabels = {
    [SdsDataSourceType.CDS]: 'CONNECT data services',
    [SdsDataSourceType.EDS]: 'Edge Data Store',
  };

  const typeOptions = [
    { value: SdsDataSourceType.CDS, label: typeLabels[SdsDataSourceType.CDS] },
    { value: SdsDataSourceType.EDS, label: typeLabels[SdsDataSourceType.EDS] },
  ];

  const edsNamespaceOptions = [
    { value: 'default', label: 'default' },
    { value: 'diagnostics', label: 'diagnostics' },
  ];

  const warningStyle = {
    color: 'orange',
    alignSelf: 'center',
  };

  const onResetClientSecret = (event: SyntheticEvent) => {
    event.preventDefault();
    const { onOptionsChange, options } = props;
    const secureJsonData = {
      ...options.secureJsonData,
      clientSecret: '',
    };
    const secureJsonFields = {
      ...options.secureJsonFields,
      clientSecret: false,
    };
    onOptionsChange({ ...options, secureJsonData, secureJsonFields });
  };

  const { options } = props;
  const { jsonData, secureJsonData } = options;

  // Fill in defaults
  if (!jsonData.type) {
    jsonData.type = SdsDataSourceType.CDS;
  }
  if (!jsonData.edsPort) {
    jsonData.edsPort = '5590';
  }
  if (!jsonData.resource) {
    jsonData.resource = 'https://int.platform.capdev-connect.aveva.com';
  }
  if (!jsonData.apiVersion) {
    jsonData.apiVersion = 'v1';
  }
  if (jsonData.oauthPassThru == null) {
    jsonData.oauthPassThru = false;
  }
  if (
    jsonData.type === SdsDataSourceType.EDS &&
    (!jsonData.sdsId || edsNamespaceOptions.findIndex((x) => x.value === jsonData.sdsId) === -1)
  ) {
    jsonData.sdsId = 'default';
  }

  return (
    <div>
      <div className="gf-form-group">
        <h3 className="page-heading">Sequential Data Store</h3>
        <div>
          <InlineField
            label="Type"
            tooltip="The type of SDS source system in use, either CONNECT data services or Edge Data Store"
            labelWidth={20}
          >
            <Select
              width={40}
              placeholder="Select type of source system..."
              options={typeOptions}
              onChange={onUpdateDatasourceJsonDataOptionSelect(props, 'type')}
              value={{ value: jsonData.type, label: typeLabels[jsonData.type] }}
            />
          </InlineField>
        </div>
      </div>
      {jsonData.type === SdsDataSourceType.EDS ? (
        <div className="gf-form-group">
          <h3 className="page-heading">Edge Data Store</h3>
          <div>
            <InlineField label="Port" tooltip="The port number used by Edge Data Store" labelWidth={20}>
              <Input
                required={true}
                placeholder="5590"
                width={40}
                onChange={onUpdateDatasourceJsonDataOption(props, 'edsPort')}
                value={jsonData.edsPort || ''}
              />
            </InlineField>
          </div>
          <div>
            <InlineField label="EDS ID" tooltip="The ID for your EDS instance" labelWidth={20}>
              <Select
                width={40}
                placeholder="EDS ID"
                options={edsNamespaceOptions}
                onChange={onUpdateDatasourceJsonDataOptionSelect(props, 'sdsId')}
                value={{ value: jsonData.sdsId, label: jsonData.sdsId }}
              />
            </InlineField>
          </div>
        </div>
      ) : (
        <div className="gf-form-group">
          <h3 className="page-heading">CONNECT data services</h3>
          <InlineField label="URL" tooltip="The URL for CONNECT" labelWidth={20}>
            <Input
              required={true}
              placeholder="https://int.platform.capdev-connect.aveva.com"
              width={40}
              onChange={onUpdateDatasourceJsonDataOption(props, 'resource')}
              value={jsonData.resource || ''}
            />
          </InlineField>
          <InlineField label="API Version" tooltip="The version of the CONNECT data services API to use" labelWidth={20}>
            <Input
              required={true}
              placeholder="v1"
              width={40}
              onChange={onUpdateDatasourceJsonDataOption(props, 'apiVersion')}
              value={jsonData.apiVersion || ''}
            />
          </InlineField>
          <InlineField label="Account ID" tooltip="The ID for your CONNECT account" labelWidth={20}>
            <Input
              required={true}
              placeholder="00000000-0000-0000-0000-000000000000"
              width={40}
              onChange={onUpdateDatasourceJsonDataOption(props, 'accountId')}
              value={jsonData.accountId || ''}
            />
          </InlineField>
          <InlineField label="SDS ID" tooltip="The Id for your SDS instance" labelWidth={20}>
            <Input
              required={true}
              placeholder="mySdsId"
              width={40}
              onChange={onUpdateDatasourceJsonDataOption(props, 'sdsId')}
              value={jsonData.sdsId || ''}
            />
          </InlineField>
          <InlineFieldRow>
            <InlineField label="Use OAuth token" tooltip="Switch to toggle authentication modes" labelWidth={20}>
              <InlineSwitch
                onChange={onUpdateDatasourceJsonDataOptionChecked(props, 'oauthPassThru')}
                value={jsonData.oauthPassThru}
              />
            </InlineField>
            {jsonData.oauthPassThru && (
              <div style={warningStyle}>
                Warning: Requires configuring genenric OAuth with CONNECT data services in your Grafana Server
              </div>
            )}
          </InlineFieldRow>
          {!jsonData.oauthPassThru && (
            <InlineField
              label="Client ID"
              tooltip="The ID of the Client Credentials client to authenticate against your ADH tenant"
              labelWidth={20}
            >
              <Input
                placeholder="00000000-0000-0000-0000-000000000000"
                width={40}
                onChange={onUpdateDatasourceJsonDataOption(props, 'clientId')}
                value={jsonData.clientId || ''}
              />
            </InlineField>
          )}
          {!jsonData.oauthPassThru && (
            <InlineField
              label="Client Secret"
              tooltip="The secret for the specified Client Credentials client"
              labelWidth={20}
            >
              <Input
                required={true}
                type="password"
                placeholder="Enter a Client secret..."
                width={40}
                onChange={onUpdateDatasourceSecureJsonDataOption(props, 'clientSecret')}
                onReset={onResetClientSecret}
                value={secureJsonData?.clientSecret || ''}
              />
            </InlineField>
          )}
        </div>
      )}
    </div>
  );
};
