/**
 * @generated SignedSource<<bd1f2132960c1cdd84906d7de58765e9>>
 * @lightSyntaxTransform
 * @nogrep
 */

/* tslint:disable */
/* eslint-disable */
// @ts-nocheck

import { ConcreteRequest } from 'relay-runtime';
export type AssignTaskInput = {
  assignedToId: string;
  taskId: string;
};
export type ControlOverviewPageAssignTaskMutation$variables = {
  input: AssignTaskInput;
};
export type ControlOverviewPageAssignTaskMutation$data = {
  readonly assignTask: {
    readonly task: {
      readonly assignedTo: {
        readonly fullName: string;
        readonly id: string;
        readonly primaryEmailAddress: string;
      } | null | undefined;
      readonly id: string;
      readonly version: number;
    };
  };
};
export type ControlOverviewPageAssignTaskMutation = {
  response: ControlOverviewPageAssignTaskMutation$data;
  variables: ControlOverviewPageAssignTaskMutation$variables;
};

const node: ConcreteRequest = (function(){
var v0 = [
  {
    "defaultValue": null,
    "kind": "LocalArgument",
    "name": "input"
  }
],
v1 = {
  "alias": null,
  "args": null,
  "kind": "ScalarField",
  "name": "id",
  "storageKey": null
},
v2 = [
  {
    "alias": null,
    "args": [
      {
        "kind": "Variable",
        "name": "input",
        "variableName": "input"
      }
    ],
    "concreteType": "AssignTaskPayload",
    "kind": "LinkedField",
    "name": "assignTask",
    "plural": false,
    "selections": [
      {
        "alias": null,
        "args": null,
        "concreteType": "Task",
        "kind": "LinkedField",
        "name": "task",
        "plural": false,
        "selections": [
          (v1/*: any*/),
          {
            "alias": null,
            "args": null,
            "kind": "ScalarField",
            "name": "version",
            "storageKey": null
          },
          {
            "alias": null,
            "args": null,
            "concreteType": "People",
            "kind": "LinkedField",
            "name": "assignedTo",
            "plural": false,
            "selections": [
              (v1/*: any*/),
              {
                "alias": null,
                "args": null,
                "kind": "ScalarField",
                "name": "fullName",
                "storageKey": null
              },
              {
                "alias": null,
                "args": null,
                "kind": "ScalarField",
                "name": "primaryEmailAddress",
                "storageKey": null
              }
            ],
            "storageKey": null
          }
        ],
        "storageKey": null
      }
    ],
    "storageKey": null
  }
];
return {
  "fragment": {
    "argumentDefinitions": (v0/*: any*/),
    "kind": "Fragment",
    "metadata": null,
    "name": "ControlOverviewPageAssignTaskMutation",
    "selections": (v2/*: any*/),
    "type": "Mutation",
    "abstractKey": null
  },
  "kind": "Request",
  "operation": {
    "argumentDefinitions": (v0/*: any*/),
    "kind": "Operation",
    "name": "ControlOverviewPageAssignTaskMutation",
    "selections": (v2/*: any*/)
  },
  "params": {
    "cacheID": "2a75b40af9d9853b851d153595dc1317",
    "id": null,
    "metadata": {},
    "name": "ControlOverviewPageAssignTaskMutation",
    "operationKind": "mutation",
    "text": "mutation ControlOverviewPageAssignTaskMutation(\n  $input: AssignTaskInput!\n) {\n  assignTask(input: $input) {\n    task {\n      id\n      version\n      assignedTo {\n        id\n        fullName\n        primaryEmailAddress\n      }\n    }\n  }\n}\n"
  }
};
})();

(node as any).hash = "02f3b1fb21bc39cea7fde9cc3945adc6";

export default node;
