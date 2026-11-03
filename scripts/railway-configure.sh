#!/usr/bin/env sh
# Configure current Railway API settings without deprecated Config as Code paths.
# Does not read, upload or display secrets. Run before publishing existing services.
set -eu
: "${RAILWAY_PROJECT_ID:?Set the existing Scooby project ID}"
: "${RAILWAY_ENVIRONMENT:?Set the environment ID}"
: "${RAILWAY_WEB_SERVICE:?Set the web service ID}"
: "${RAILWAY_ENGINE_SERVICE:?Set the engine service ID}"
command -v railway >/dev/null
command -v node >/dev/null
for role in web engine; do
  ROLE="$role" node -e '
    const role=process.env.ROLE;
    const input={dockerfilePath:"Dockerfile",sleepApplication:false,numReplicas:1,
      startCommand:role==="web"?"/app/app":"/app/engine",healthcheckPath:"/ready",
      healthcheckTimeout:120,restartPolicyType:"ON_FAILURE",restartPolicyMaxRetries:role==="web"?5:10};
    if(role==="web")Object.assign(input,{preDeployCommand:["/app/app migrate"],preDeployTimeoutSeconds:120});
    process.stdout.write(JSON.stringify({service:role==="web"?process.env.RAILWAY_WEB_SERVICE:process.env.RAILWAY_ENGINE_SERVICE,
      env:process.env.RAILWAY_ENVIRONMENT,input}));
  ' | railway api 'mutation($service:String!,$env:String!,$input:ServiceInstanceUpdateInput!){serviceInstanceUpdate(serviceId:$service,environmentId:$env,input:$input)}' --variables @- --compact
done
