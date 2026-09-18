import nextCoreWebVitals from "eslint-config-next/core-web-vitals";
import nextTypescript from "eslint-config-next/typescript";

const eslintConfig = [
  ...nextCoreWebVitals,
  ...nextTypescript,
  {
    // eslint-plugin-react's automatic React-version detection calls the
    // removed ESLint 9+ `context.getFilename()` API and crashes under
    // ESLint 10. Setting the version explicitly skips that code path.
    settings: { react: { version: "19.2" } },
  },
  {
    ignores: [".next/**", "dist/**", "node_modules/**"],
  },
];

export default eslintConfig;
