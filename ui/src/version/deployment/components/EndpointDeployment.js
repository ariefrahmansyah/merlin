/**
 * Copyright 2020 The Merlin Authors
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *      http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

import React, { Fragment, useCallback, useEffect, useState } from "react";
import { navigate } from "@reach/router";
import {
  EuiButton,
  EuiButtonEmpty,
  EuiConfirmModal,
  EuiFlexGroup,
  EuiFlexItem,
  EuiForm,
  EuiOverlayMask,
  EuiProgress,
  EuiSpacer
} from "@elastic/eui";
import { replaceBreadcrumbs, useToggle } from "@gojek/mlp-ui";
import { useMerlinApi } from "../../../hooks/useMerlinApi";
import mocks from "../../../mocks";
import { EndpointEnvironment } from "./EndpointEnvironment";
import { EndpointResources } from "./EndpointResources";
import { EndpointVariables } from "./EndpointVariables";

const DeployConfirmationModal = ({
  actionTitle,
  content,
  isLoading,
  onConfirm,
  onCancel
}) => (
  <EuiOverlayMask>
    <EuiConfirmModal
      title={`${actionTitle} model version`}
      onCancel={onCancel}
      onConfirm={onConfirm}
      cancelButtonText="Cancel"
      confirmButtonText={actionTitle}>
      {content}
      {isLoading && <EuiProgress size="xs" color="accent" />}
    </EuiConfirmModal>
  </EuiOverlayMask>
);

const defaultResourceRequest = {
  cpu_request: "500m",
  memory_request: "500Mi",
  min_replica: process.env.REACT_APP_ENVIRONMENT === "production" ? 2 : 0,
  max_replica: process.env.REACT_APP_ENVIRONMENT === "production" ? 4 : 2
};

const targetRequestStatus = currentStatus => {
  return currentStatus === "serving" ? "serving" : "running";
};

export const EndpointDeployment = ({
  actionTitle,
  breadcrumbs,
  model,
  version,
  endpointId,
  disableEnvironment,
  modalContent,
  onDeploy,
  response
}) => {
  useEffect(() => {
    replaceBreadcrumbs([...breadcrumbs, { text: actionTitle }]);
  }, [actionTitle, breadcrumbs]);

  const redirectUrl = `/merlin/projects/${model.project_id}/models/${model.id}/versions`;

  const [request, setRequest] = useState({});

  useEffect(() => {
    version.endpoints &&
      endpointId &&
      setRequest(version.endpoints.find(e => e.id === endpointId));
  }, [version.endpoints, endpointId, setRequest]);

  const [{ data: environments }] = useMerlinApi(
    `/environments`,
    { mock: mocks.environmentList },
    [],
    true
  );

  const [isModalVisible, toggleModalVisible] = useToggle();

  const openModal = () => {
    if (!isModalVisible && response.isLoaded) {
      response.isLoaded = false;
    }
    toggleModalVisible();
  };

  useEffect(() => {
    if (response.isLoaded) {
      isModalVisible && toggleModalVisible();
      if (!response.error) {
        navigate(redirectUrl);
      }
    }
  }, [response, isModalVisible, toggleModalVisible, redirectUrl]);

  // first, get selected environment's default resource request
  // next, check if endpoint in selected environment already exists or not
  // if exists and status is not pending, it's a redeployment so let's get previous configurations
  useEffect(() => {
    const selectedEnvironment = environments.find(
      e => e.name === request.environment_name
    );

    let targetResourceRequest = defaultResourceRequest;
    if (selectedEnvironment) {
      if (selectedEnvironment.default_resource_request) {
        targetResourceRequest = selectedEnvironment.default_resource_request;
      }
    }

    const endpoint = version.endpoints.find(
      e => e.environment_name === request.environment_name
    );

    let targetEnvVars = [];
    if (endpoint) {
      if (endpoint.resource_request) {
        targetResourceRequest = endpoint.resource_request;
      }
      if (endpoint.env_vars) {
        targetEnvVars = endpoint.env_vars.filter(
          envVar => envVar.name !== "MODEL_NAME" && envVar.name !== "MODEL_DIR"
        );
      }
    }

    setRequest(r => ({
      ...r,
      resource_request: targetResourceRequest,
      env_vars: targetEnvVars
    }));
  }, [environments, version.endpoints, request.environment_name, setRequest]);

  const onChange = field => {
    return value => setRequest(r => ({ ...r, [field]: value }));
  };

  const onVariablesChange = useCallback(onChange("env_vars"), []);

  return (
    <Fragment>
      <EuiSpacer size="l" />

      <EuiFlexGroup justifyContent="spaceAround">
        <EuiFlexItem style={{ maxWidth: 600 }}>
          <EuiForm
            isInvalid={!!response.error}
            error={response.error ? [response.error.message] : ""}>
            <EuiFlexGroup direction="column">
              <EuiFlexItem grow={false}>
                <EndpointEnvironment
                  version={version}
                  selected={request.environment_name}
                  environments={environments}
                  disabled={disableEnvironment}
                  onChange={onChange("environment_name")}
                />
              </EuiFlexItem>

              <EuiFlexItem grow={false}>
                <EuiSpacer size="s" />
                <EndpointResources
                  resourceRequest={
                    request.resource_request || defaultResourceRequest
                  }
                  onChange={onChange("resource_request")}
                />
              </EuiFlexItem>

              {model.type === "pyfunc" && (
                <EuiFlexItem grow={false}>
                  <EuiSpacer size="s" />
                  <EndpointVariables
                    variables={request.env_vars || []}
                    onChange={onVariablesChange}
                  />
                </EuiFlexItem>
              )}

              <EuiFlexItem grow={false}>
                <EuiFlexGroup direction="row" justifyContent="flexEnd">
                  <EuiFlexItem grow={false}>
                    <EuiButtonEmpty
                      size="s"
                      onClick={() => navigate(redirectUrl)}>
                      Cancel
                    </EuiButtonEmpty>
                  </EuiFlexItem>

                  <EuiFlexItem grow={false}>
                    <EuiButton
                      size="s"
                      color="primary"
                      fill
                      disabled={!request.environment_name}
                      onClick={openModal}>
                      {actionTitle}
                    </EuiButton>
                  </EuiFlexItem>
                </EuiFlexGroup>
              </EuiFlexItem>
            </EuiFlexGroup>
          </EuiForm>
        </EuiFlexItem>

        {isModalVisible && (
          <DeployConfirmationModal
            actionTitle={actionTitle}
            content={modalContent}
            isLoading={response.isLoading}
            onConfirm={() =>
              onDeploy({
                body: JSON.stringify({
                  ...request,
                  status: targetRequestStatus(request.status)
                })
              })
            }
            onCancel={toggleModalVisible}
          />
        )}
      </EuiFlexGroup>
    </Fragment>
  );
};
