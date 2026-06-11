# Requirements Document

## Introduction

Ghost loads its provider and model list during configuration `Load()`. For the OpenCode Zen provider, the available models are read from a local cache (cache key `dynamic_providers`) with a fallback to a static embedded list. A live fetch from the OpenCode Zen model endpoint (`https://opencode.ai/zen/v1/models`) and a corresponding cache write exist in code (`fetchFreshDynamicProviders` and `saveDynamicProvidersToCache`), but no part of the application triggers them on or after startup. As a result, the OpenCode Zen model list becomes stale: newly added free models do not appear and removed models continue to be offered.

This feature adds an automatic post-startup refresh of the OpenCode Zen model list. The application starts using the cache-first/static list so startup stays fast and non-blocking, then refreshes the OpenCode Zen models from the live endpoint in the background and updates the local cache so subsequent sessions reflect the current model list. The refresh honors the existing auto-update disable option (`DisableProviderAutoUpdate`, environment variable `GHOST_DISABLE_PROVIDER_AUTO_UPDATE`).

## Glossary

- **Application**: The Ghost CLI/TUI program that loads configuration and providers at startup.
- **Model_Refresh_Service**: The component responsible for refreshing the OpenCode Zen model list after startup and updating the local cache.
- **OpenCode_Zen_Endpoint**: The remote HTTP endpoint at `https://opencode.ai/zen/v1/models` that returns the authoritative OpenCode Zen model list.
- **Dynamic_Providers_Cache**: The local on-disk cache identified by cache key `dynamic_providers` that stores the OpenCode Zen provider and its model list.
- **Free_Model**: An OpenCode Zen model whose ID is suffixed with `-free`, plus the stealth model with ID `big-pickle`.
- **Static_Fallback_List**: The embedded list of OpenCode Zen models returned when the cache is empty or invalid.
- **Auto_Update_Disabled**: The configuration state where `Options.DisableProviderAutoUpdate` is `true` (set directly or via environment variable `GHOST_DISABLE_PROVIDER_AUTO_UPDATE`).
- **Refresh_Cycle**: A single attempt by the Model_Refresh_Service to fetch the live model list, filter it, and update the Dynamic_Providers_Cache.
- **Startup**: The phase during which the Application loads configuration and providers and becomes ready for user interaction.

## Requirements

### Requirement 1: Automatic post-startup refresh

**User Story:** As a Ghost user, I want the OpenCode Zen model list to refresh automatically after the app starts, so that I always see the current set of available free models without running a manual command.

#### Acceptance Criteria

1. WHEN Startup completes and Auto_Update_Disabled is false, THE Model_Refresh_Service SHALL initiate exactly one Refresh_Cycle within 5 seconds of Startup completion.
2. WHILE a Refresh_Cycle is running, THE Application SHALL accept and process user input without blocking on the Refresh_Cycle and SHALL serve the model list that was loaded during Startup.
3. WHEN a Refresh_Cycle completes successfully, THE Model_Refresh_Service SHALL write the fetched OpenCode Zen provider and model list to the Dynamic_Providers_Cache before the Refresh_Cycle is reported as complete.
4. WHERE Auto_Update_Disabled is true, THE Model_Refresh_Service SHALL NOT initiate a Refresh_Cycle.
5. IF a Refresh_Cycle is already running when Startup completes, THEN THE Model_Refresh_Service SHALL NOT initiate an additional concurrent Refresh_Cycle.

### Requirement 2: Non-blocking startup

**User Story:** As a Ghost user, I want startup to stay fast, so that I can begin working immediately even when the network is slow or unavailable.

#### Acceptance Criteria

1. WHEN the Application starts, THE Application SHALL load the OpenCode Zen model list from the Dynamic_Providers_Cache when the cache contains a valid entry, where a valid entry is one that is non-empty and parses successfully.
2. IF the Dynamic_Providers_Cache is empty or unparseable, THEN THE Application SHALL load the Static_Fallback_List and SHALL complete Startup.
3. THE Model_Refresh_Service SHALL perform each Refresh_Cycle on a background path that runs concurrently with and separate from the Startup loading path, such that the Startup loading path issues no network request to the OpenCode_Zen_Endpoint.
4. WHEN a Refresh_Cycle is initiated, THE Application SHALL complete Startup within 2000 milliseconds without waiting for the Refresh_Cycle to finish.

### Requirement 3: Live fetch and free-model filtering

**User Story:** As a Ghost user, I want only valid free OpenCode Zen models to be stored, so that the model list stays accurate and usable.

#### Acceptance Criteria

1. WHEN a Refresh_Cycle runs, THE Model_Refresh_Service SHALL request the model list from the OpenCode_Zen_Endpoint.
2. WHEN the OpenCode_Zen_Endpoint returns a response with a success status, THE Model_Refresh_Service SHALL include in the refreshed model list every returned model whose model ID ends with the suffix `-free` or whose model ID equals `big-pickle`.
3. WHEN the OpenCode_Zen_Endpoint returns a model whose model ID neither ends with the suffix `-free` nor equals `big-pickle`, THE Model_Refresh_Service SHALL exclude that model from the refreshed model list.
4. WHEN two or more entries in the OpenCode_Zen_Endpoint response share an identical model ID, THE Model_Refresh_Service SHALL retain only the first such entry in endpoint response order and SHALL discard each subsequent entry with that model ID.
5. WHEN the OpenCode_Zen_Endpoint returns a Free_Model that is absent from the Dynamic_Providers_Cache, THE Model_Refresh_Service SHALL add that model to the refreshed model list.
6. WHEN a model present in the Dynamic_Providers_Cache is absent from the OpenCode_Zen_Endpoint response, THE Model_Refresh_Service SHALL exclude that model from the refreshed model list.

### Requirement 4: Failure handling

**User Story:** As a Ghost user, I want the app to keep working when the refresh fails, so that a network or endpoint problem never breaks my session or corrupts my model list.

#### Acceptance Criteria

1. IF the request to the OpenCode_Zen_Endpoint fails, THEN THE Model_Refresh_Service SHALL retain the existing Dynamic_Providers_Cache contents unchanged.
2. IF the OpenCode_Zen_Endpoint returns a non-success HTTP status, THEN THE Model_Refresh_Service SHALL retain the existing Dynamic_Providers_Cache contents unchanged.
3. IF the OpenCode_Zen_Endpoint response cannot be decoded, THEN THE Model_Refresh_Service SHALL retain the existing Dynamic_Providers_Cache contents unchanged.
4. IF a Refresh_Cycle produces zero Free_Model entries, THEN THE Model_Refresh_Service SHALL retain the existing Dynamic_Providers_Cache contents unchanged.
5. IF a Refresh_Cycle fails, THEN THE Model_Refresh_Service SHALL record one diagnostic log entry that identifies the failure cause as exactly one of: request error, non-success HTTP status, decode failure, zero Free_Model entries, or 15-second timeout.
6. WHILE the request to the OpenCode_Zen_Endpoint has been running for more than 15 seconds, THE Model_Refresh_Service SHALL abort the request and treat the Refresh_Cycle as failed.
7. WHEN the request to the OpenCode_Zen_Endpoint completes at or before 15 seconds, THE Model_Refresh_Service SHALL allow the request to finish without aborting it.
8. WHEN a Refresh_Cycle fails, THE Model_Refresh_Service SHALL return control to the Application without raising an unhandled error, leaving the Application able to continue operating using the retained Dynamic_Providers_Cache contents.

### Requirement 5: Cache persistence and reuse

**User Story:** As a Ghost user, I want refreshed models to be remembered between runs, so that the updated list appears on my next session without another network call at startup.

#### Acceptance Criteria

1. WHEN a Refresh_Cycle completes successfully, THE Model_Refresh_Service SHALL persist the refreshed OpenCode Zen provider and model list to the Dynamic_Providers_Cache under cache key `dynamic_providers`.
2. IF persisting the refreshed model list to the Dynamic_Providers_Cache fails, THEN THE Model_Refresh_Service SHALL retain the previously stored cache entry unchanged and surface an error indication that the cache write did not complete.
3. WHEN the Application starts and the Dynamic_Providers_Cache contains an entry under cache key `dynamic_providers`, THE Application SHALL load the cached OpenCode Zen model list without performing a startup network call to refresh that list.
4. WHEN the Model_Refresh_Service writes to the Dynamic_Providers_Cache, THE Model_Refresh_Service SHALL replace the entire stored OpenCode Zen model list with the refreshed model list, removing any entries not present in the refreshed list.
5. WHEN the Application starts and the Dynamic_Providers_Cache contains an entry under cache key `dynamic_providers`, THE Application SHALL load that cached model list irrespective of the elapsed time since the entry was written.
6. IF the Application starts and the Dynamic_Providers_Cache contains no entry under cache key `dynamic_providers` or the stored entry cannot be read or parsed, THEN THE Application SHALL proceed without a cached model list and SHALL NOT terminate Startup due to the absent or unreadable entry.
