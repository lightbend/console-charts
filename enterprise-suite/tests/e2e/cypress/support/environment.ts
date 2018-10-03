import { environment as LocalEnv } from './environment.dev';
import { environment as MinikubeEnv } from './environment.prod';
import { Env } from './environment.base';

export class Environment {
  static getEnv(): Env {
    const env: 'minikube' | 'local' = Cypress.env().configFile;
    const defaultEnv = LocalEnv;
    return (env === 'minikube') ? MinikubeEnv : defaultEnv;
  }
}
