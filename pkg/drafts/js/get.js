let d = Draft.find(input[0]);
if (d) {
  let folder = d.isTrashed ? "trash" : (d.isArchived ? "archive" : "inbox");
  let res = {
    uuid: d.uuid,
    content: d.content,
    title: d.displayTitle,
    tags: d.tags,
    isFlagged: d.isFlagged,
    isArchived: d.isArchived,
    isTrashed: d.isTrashed,
    folder: folder,
    createdAt: d.createdAt.toISOString(),
    modifiedAt: d.modifiedAt.toISOString(),
    permalink: d.permalink
  };
  context.addSuccessParameter("result", JSON.stringify(res));
}
