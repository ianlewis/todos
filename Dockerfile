# Copyright 2025 Ian Lewis
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#      http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.

# This Dockerfile uses multi-platform builds but the cross-compilation happens
# externally to Docker prior to running `docker build`.
FROM --platform=$BUILDPLATFORM scratch
ARG TARGETOS
ARG TARGETARCH

COPY --chmod=0755 todos-${TARGETOS}-${TARGETARCH} /todos

ENTRYPOINT ["/todos"]
