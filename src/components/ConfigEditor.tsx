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
    [SdsDataSourceType.ADH]: 'CONNECT data services',
    [SdsDataSourceType.EDS]: 'Edge Data Store',
  };

  const typeOptions = [
    { value: SdsDataSourceType.ADH, label: typeLabels[SdsDataSourceType.ADH] },
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
    jsonData.type = SdsDataSourceType.ADH;
  }
  if (!jsonData.edsPort) {
    jsonData.edsPort = '5590';
  }
  if (!jsonData.resource) {
    jsonData.resource = 'https://uswe.datahub.connect.aveva.com';
  }
  if (!jsonData.apiVersion) {
    jsonData.apiVersion = 'v1';
  }
  if (jsonData.oauthPassThru == null) {
    jsonData.oauthPassThru = false;
  }
  if (
    jsonData.type === SdsDataSourceType.EDS &&
    (!jsonData.namespaceId || edsNamespaceOptions.findIndex((x) => x.value === jsonData.namespaceId) === -1)
  ) {
    jsonData.namespaceId = 'default';
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
            <InlineField label="Namespace" tooltip="The Namespace in your for CONNECT data services tenant" labelWidth={20}>
              <Select
                width={40}
                placeholder="EDS Namespace"
                options={edsNamespaceOptions}
                onChange={onUpdateDatasourceJsonDataOptionSelect(props, 'namespaceId')}
                value={{ value: jsonData.namespaceId, label: jsonData.namespaceId }}
              />
            </InlineField>
          </div>
        </div>
      ) : (
        <div className="gf-form-group">
          <h3 className="page-heading">CONNECT data services</h3>
          <InlineField label="URL" tooltip="The URL for CONNECT data services" labelWidth={20}>
            <Input
              required={true}
              placeholder="https://uswe.datahub.connect.aveva.com"
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
          <InlineField label="Tenant ID" tooltip="The ID of your CONNECT data services tenant" labelWidth={20}>
            <Input
              required={true}
              placeholder="00000000-0000-0000-0000-000000000000"
              width={40}
              onChange={onUpdateDatasourceJsonDataOption(props, 'tenantId')}
              value={jsonData.tenantId || ''}
            />
          </InlineField>
          <InlineFieldRow>
            <InlineField
              label="Community Data"
              tooltip="Switch to toggle reading from a Namespace to a Community"
              labelWidth={20}
            >
              <InlineSwitch
                onChange={onUpdateDatasourceJsonDataOptionChecked(props, 'useCommunity')}
                value={jsonData.useCommunity}
              />
            </InlineField>
          </InlineFieldRow>
          {jsonData.useCommunity && (
            <InlineField label="Community ID" tooltip="The ID of the CONNECT data services Community" labelWidth={20}>
              <Input
                placeholder="00000000-0000-0000-0000-000000000000"
                width={40}
                onChange={onUpdateDatasourceJsonDataOption(props, 'communityId')}
                value={jsonData.communityId || ''}
              />
            </InlineField>
          )}
          {!jsonData.useCommunity && (
            <InlineField label="Namespace ID" tooltip="The Namespace in your for CONNECT data services tenant" labelWidth={20}>
              <Input
                required={true}
                placeholder="00000000-0000-0000-0000-000000000000"
                width={40}
                onChange={onUpdateDatasourceJsonDataOption(props, 'namespaceId')}
                value={jsonData.namespaceId || ''}
              />
            </InlineField>
          )}
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
              tooltip="The ID of the Client Credentials client to authenticate against your Cds tenant"
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
