import React from 'react';
import { AsyncSelect, InlineFormLabel } from '@grafana/ui';
import { QueryEditorProps, SelectableValue } from '@grafana/data';
import { DataSource } from '../datasource';
import { defaultQuery, SdsDataSourceOptions, SdsQuery } from '../types';
import { debounce } from '../debounce';

type Props = QueryEditorProps<DataSource, SdsQuery, SdsDataSourceOptions>;

export function QueryEditor({ query, datasource, onChange }: Props) {
  const combinedQuery = { ...defaultQuery, ...query };

  const selectStream: SelectableValue<string> = { label: combinedQuery.name, value: combinedQuery.id };
  const [defaultOptions, setDefaultOptions] = React.useState<boolean | Array<SelectableValue<string>>>(true);

  const onSelectedStream = (value: SelectableValue<string>) => {
    onChange({ ...combinedQuery, id: value.value || '', name: value.label || '' });
  };

  const debouncedGetStreams = debounce(
    (inputvalue: string) => datasource.getStreams(inputvalue, setDefaultOptions),
    1000
  );

return (
    <div className="gf-form">
      <InlineFormLabel width={8}>Stream</InlineFormLabel>
      <AsyncSelect
        defaultOptions={defaultOptions}
        width={50}
        loadOptions={debouncedGetStreams}
        value={selectStream}
        onChange={onSelectedStream}
        placeholder="Select Stream"
        loadingMessage={'Loading streams...'}
        noOptionsMessage={'No streams found'}
      />
    </div>
  );
}
