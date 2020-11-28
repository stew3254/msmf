const username = $("#username");
const password = $("#password");
const register = $("button");
const code = $("#code");

function submit() {
  $.ajax({
    url : "/api/refer/"+code.val(),
    type: "POST",
    data : JSON.stringify({
      username: username.val(),
      password: password.val(),
    }),
    success: function(data, textStatus, jqXHR) {
      alert("Success");
    },
    error: function (jqXHR, textStatus, errorThrown) {
      console.log("failture:", errorThrown);
    }
  });
}

register.on("click", () => {
  submit();
})