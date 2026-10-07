from collections.abc import Callable
from http import HTTPStatus

import requests

from fixtures import types
from fixtures.auth import USER_ADMIN_EMAIL, USER_ADMIN_PASSWORD

MATERIALIZED_PATH = "/api/v1/promoted_path/materialized"
PROBE_INDEX_EXPRS_QUERY = "SELECT expr FROM system.data_skipping_indices WHERE database = 'signoz_traces' AND table = 'signoz_index_v3' AND expr LIKE 'CAST(%materialized.%probe%'"


def test_index_materialized_paths(
    signoz: types.SigNoz,
    create_user_admin: types.Operation,  # pylint: disable=unused-argument
    get_token: Callable[[str, str], str],
    materialize_attribute: Callable[[str, str, bool], str],
) -> None:
    materialize_attribute("traces", "materialized.probe", True)
    materialize_attribute("traces", "materialized.unindexed_probe", False)
    token = get_token(USER_ADMIN_EMAIL, USER_ADMIN_PASSWORD)

    resp = requests.post(
        signoz.self.host_configs["8080"].get(MATERIALIZED_PATH),
        params={"dryRun": "true"},
        headers={"Authorization": f"Bearer {token}"},
        timeout=30,
    )
    assert resp.status_code == HTTPStatus.OK, resp.text
    indexed = {path["path"]: path["indexes"] for path in resp.json()["data"]["indexed"]}
    skipped = {path["path"]: path["reason"] for path in resp.json()["data"]["skipped"]}
    assert indexed["materialized.probe"] == [{"fieldDataType": "string", "type": "bloom_filter(0.01)", "granularity": 64}]
    assert skipped["materialized.unindexed_probe"] == "column_not_indexed"
    assert skipped["http.route"] == "default_materialized"
    assert not signoz.telemetrystore.conn.query(PROBE_INDEX_EXPRS_QUERY).result_rows

    resp = requests.post(
        signoz.self.host_configs["8080"].get(MATERIALIZED_PATH),
        params={"signal": "logs"},
        headers={"Authorization": f"Bearer {token}"},
        timeout=30,
    )
    assert resp.status_code == HTTPStatus.OK, resp.text
    assert resp.json()["data"] == {"indexed": [], "skipped": []}
    assert not signoz.telemetrystore.conn.query(PROBE_INDEX_EXPRS_QUERY).result_rows

    resp = requests.post(
        signoz.self.host_configs["8080"].get(MATERIALIZED_PATH),
        headers={"Authorization": f"Bearer {token}"},
        timeout=30,
    )
    assert resp.status_code == HTTPStatus.OK, resp.text
    assert "materialized.probe" in {path["path"] for path in resp.json()["data"]["indexed"]}
    assert [row[0] for row in signoz.telemetrystore.conn.query(PROBE_INDEX_EXPRS_QUERY).result_rows] == ["CAST(attributes.materialized.probe, 'String')"]

    resp = requests.post(
        signoz.self.host_configs["8080"].get(MATERIALIZED_PATH),
        headers={"Authorization": f"Bearer {token}"},
        timeout=30,
    )
    assert resp.status_code == HTTPStatus.OK, resp.text
    assert "materialized.probe" not in {path["path"] for path in resp.json()["data"]["indexed"]}
    assert {path["path"]: path["reason"] for path in resp.json()["data"]["skipped"]}["materialized.probe"] == "already_indexed"
