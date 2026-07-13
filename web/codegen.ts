import type { CodegenConfig } from "@graphql-codegen/cli";

const config: CodegenConfig = {
  schema: "../src/api-gateway/graph/schemas/*.graphqls",
  documents: "src/**/*.graphql",
  generates: {
    "src/gql/": {
      preset: "client",
      config: {
        useTypeImports: true,
        scalars: { Int64: { input: "number", output: "number" } },
      },
    },
  },
};

export default config;
