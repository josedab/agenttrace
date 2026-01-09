import { URLExt } from '@jupyterlab/coreutils';
import { ServerConnection } from '@jupyterlab/services';

/**
 * Call the AgentTrace API endpoint.
 *
 * @param endPoint API REST end point for the extension
 * @param init Optional parameters for the fetch call
 * @returns The response body interpreted as JSON
 */
export async function requestAPI<T>(
  endPoint = '',
  init: RequestInit = {}
): Promise<T> {
  // Make request to Jupyter API
  const settings = ServerConnection.makeSettings();
  const requestUrl = URLExt.join(
    settings.baseUrl,
    'agenttrace',
    endPoint
  );

  let response: Response;
  try {
    response = await ServerConnection.makeRequest(requestUrl, init, settings);
  } catch (error) {
    throw new ServerConnection.NetworkError(error as any);
  }

  let data: any = await response.text();

  if (data.length > 0) {
    try {
      data = JSON.parse(data);
    } catch (error) {
      console.error('Not a JSON response body.', response);
    }
  }

  if (!response.ok) {
    throw new ServerConnection.ResponseError(response, data.message || data);
  }

  return data;
}

/**
 * Check if AgentTrace is available and configured.
 */
export async function checkAgentTraceStatus(): Promise<{
  available: boolean;
  configured: boolean;
}> {
  try {
    const config = await requestAPI<any>('config');
    return {
      available: config.available,
      configured: config.configured
    };
  } catch (error) {
    return {
      available: false,
      configured: false
    };
  }
}
