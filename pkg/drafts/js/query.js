let [queryString, filter, tags, omitTags, sort, sortDescending, sortFlaggedToTop] = input;
let ds = Draft.query(queryString, filter, tags || [], omitTags || [], sort, sortDescending, sortFlaggedToTop);
if (!ds) ds = [];
let res = ds.map((d) => ({
  uuid: d.uuid,
  content: d.content,
  tags: d.tags,
  isFlagged: d.isFlagged,
  isArchived: d.isArchived,
  isTrashed: d.isTrashed,
}));
if (ds.length > 0) context.addSuccessParameter("result", JSON.stringify(res));
