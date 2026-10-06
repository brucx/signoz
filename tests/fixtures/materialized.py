from collections.abc import Callable, Generator

import pytest

from fixtures import types

# signal -> (database, local table, distributed table)
MATERIALIZED_TABLES = {
    "traces": ("signoz_traces", "signoz_index_v3", "distributed_signoz_index_v3"),
    "logs": ("signoz_logs", "logs_v2", "distributed_logs_v2"),
}


@pytest.fixture(name="materialize_column", scope="function")
def materialize_column(signoz: types.SigNoz) -> Generator[Callable[[str, str, str, str | None], str]]:
    """
    Yields materialize(signal, context, key, index) that adds the legacy `<context>_string_<key>` column, plus `index`
    (e.g. "tokenbf_v1(1024, 2, 0) GRANULARITY 4") on it when given, and returns the column name.
    Teardown drops the columns and every index whose expression mentions a materialized key.
    """
    cluster = signoz.telemetrystore.env["SIGNOZ_TELEMETRYSTORE_CLICKHOUSE_CLUSTER"]
    created: list[tuple[str, str, str]] = []

    def materialize(signal: str, context: str, key: str, index: str | None) -> str:
        database, local_table, distributed_table = MATERIALIZED_TABLES[signal]
        map_column = {"attribute": "attributes_string", "resource": "resources_string"}[context]
        column = f"{context}_string_{key.replace('.', '$$')}"
        for table in (local_table, distributed_table):
            signoz.telemetrystore.conn.query(f"ALTER TABLE {database}.{table} ON CLUSTER '{cluster}' ADD COLUMN IF NOT EXISTS `{column}` LowCardinality(String) DEFAULT {map_column}['{key}']")
        if index:
            signoz.telemetrystore.conn.query(f"ALTER TABLE {database}.{local_table} ON CLUSTER '{cluster}' ADD INDEX IF NOT EXISTS `{column}_idx` `{column}` TYPE {index}")
        created.append((signal, key, column))
        return column

    yield materialize

    for signal, key, column in created:
        database, local_table, distributed_table = MATERIALIZED_TABLES[signal]
        indexes = signoz.telemetrystore.conn.query(f"SELECT name FROM system.data_skipping_indices WHERE database = '{database}' AND table = '{local_table}' AND (expr LIKE '%{key}%' OR expr LIKE '%{column}%')").result_rows
        for (index_name,) in indexes:
            signoz.telemetrystore.conn.query(f"ALTER TABLE {database}.{local_table} ON CLUSTER '{cluster}' DROP INDEX IF EXISTS `{index_name}`")
        for table in (local_table, distributed_table):
            signoz.telemetrystore.conn.query(f"ALTER TABLE {database}.{table} ON CLUSTER '{cluster}' DROP COLUMN IF EXISTS `{column}`")
