from collections.abc import Callable
from http import HTTPStatus

import requests

from fixtures import types
from fixtures.auth import USER_ADMIN_EMAIL, USER_ADMIN_PASSWORD

MATERIALIZED_PATH = "/api/v1/promoted_path/materialized"
PROBE_INDEX_EXPRS_QUERY = "SELECT expr FROM system.data_skipping_indices WHERE (database, table) IN (('signoz_traces', 'signoz_index_v3'), ('signoz_logs', 'logs_v2')) AND expr LIKE 'CAST(%materialized.probe%'"


def test_index_materialized_paths(
    signoz: types.SigNoz,
    create_user_admin: types.Operation,  # pylint: disable=unused-argument
    get_token: Callable[[str, str], str],
    materialize_column: Callable[[str, str, str, str | None], str],
) -> None:
    materialize_column("traces", "attribute", "materialized.probe", "tokenbf_v1(1024, 2, 0) GRANULARITY 4")
    materialize_column("logs", "resource", "materialized.probe", None)
    token = get_token(USER_ADMIN_EMAIL, USER_ADMIN_PASSWORD)

    resp = requests.post(
        signoz.self.host_configs["8080"].get(MATERIALIZED_PATH),
        params={"dryRun": "true"},
        headers={"Authorization": f"Bearer {token}"},
        timeout=30,
    )
    assert resp.status_code == HTTPStatus.OK, resp.text
    paths = {(path["signal"], path["context"], path["path"]): path["indexes"] for path in resp.json()["data"]}
    assert paths[("traces", "attribute", "materialized.probe")] == [{"fieldDataType": "string", "type": "tokenbf_v1(1024, 2, 0)", "granularity": 4}]
    assert paths[("logs", "resource", "materialized.probe")] == [{"fieldDataType": "string", "type": "bloom_filter(0.01)", "granularity": 64}]
    assert ("traces", "attribute", "http.route") not in paths
    assert ("traces", "resource", "service.name") not in paths
    assert not signoz.telemetrystore.conn.query(PROBE_INDEX_EXPRS_QUERY).result_rows

    resp = requests.post(
        signoz.self.host_configs["8080"].get(MATERIALIZED_PATH),
        headers={"Authorization": f"Bearer {token}"},
        timeout=30,
    )
    assert resp.status_code == HTTPStatus.OK, resp.text
    assert {(path["signal"], path["context"], path["path"]) for path in resp.json()["data"]} >= {("traces", "attribute", "materialized.probe"), ("logs", "resource", "materialized.probe")}
    assert {row[0] for row in signoz.telemetrystore.conn.query(PROBE_INDEX_EXPRS_QUERY).result_rows} == {
        "CAST(attributes.materialized.probe, 'String')",
        "CAST(resource.materialized.probe, 'String')",
    }

    resp = requests.post(
        signoz.self.host_configs["8080"].get(MATERIALIZED_PATH),
        headers={"Authorization": f"Bearer {token}"},
        timeout=30,
    )
    assert resp.status_code == HTTPStatus.OK, resp.text
    assert not [path for path in resp.json()["data"] if path["path"] == "materialized.probe"]
