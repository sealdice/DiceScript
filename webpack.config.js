export default {
  // …
  experiments: {
    outputModule: true,
  },
  output: {
    library: {
      // do not specify a `name` here
      type: 'module',
    },
  },
  // mode: 'development',
  // optimization: {
  //   usedExports: true,
  // },
};
