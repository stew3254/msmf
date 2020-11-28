const invite = $("#invite");

function submit() {
  return $.ajax({
    url : "/api/refer/new",
    type: "GET",
    success: function(data, textStatus, jqXHR) {
      console.log("Good")
    },
    error: function (jqXHR, textStatus, errorThrown) {
      console.log("failture:", errorThrown);
    }
  });
}

invite.on("click", () => {
  submit().then(r => {
    if (r != null)
      alert("Code: " + JSON.parse(r).code);
  })
})