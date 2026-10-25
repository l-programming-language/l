// Static focusability also works when JavaScript is disabled. Expressive Code
// can still refine focusability on resize when its own small client script runs.
export const readableCode = {
  name: 'Keyboard-readable code',
  hooks: {
    postprocessRenderedBlock({ renderData }) {
      function focusPre(node) {
        if (node.type === 'element' && node.tagName === 'pre') node.properties.tabIndex = 0;
        for (const child of node.children ?? []) focusPre(child);
      }
      focusPre(renderData.blockAst);
    },
  },
};

export function focusTables() {
  return (tree) => {
    function visit(node) {
      if (node.type === 'table') {
        node.data ??= {};
        node.data.hProperties = { ...node.data.hProperties, tabIndex: 0 };
      }
      if (node.children) node.children = node.children.flatMap((child) => child.type === 'table' && child.children?.[0]?.children?.length >= 3
        ? [{ type: 'html', value: '<p class="l-table-hint">Table: scroll horizontally to read all columns.</p>' }, child]
        : [child]);
      for (const child of node.children ?? []) visit(child);
    }
    visit(tree);
  };
}
