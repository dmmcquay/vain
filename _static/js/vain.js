$(function() {
    $("#name").val("");
    $("#name").keyup(function(e) {
        if ($(this).val() == "") {
            $("#name-holder").removeClass("has-success").addClass("has-warning");
        } else {
            $("#name-holder").removeClass("has-warning has-error").addClass("has-default");
        }
        if (e.keyCode == 13) {
            send();
        }
    })
    $("#send").click(function(e) {
        send();
    });
});

function send() {
    $("#alert").hide();
    $("#success").hide();
    var name = $("#name").val();
    if (name == "") {
        $("#name-holder").removeClass("has-warning has-default").addClass("has-error");
        $("#alert").show().text("Please provide an email address.");
    } else {
        route = "/api/v0/register/?email="+name;
        if(window.location.href.indexOf("forgot") > -1) {
            route = "/api/v0/forgot/?email="+name;
        }
        $.get(route).done(
            function(data) {
                $("#input").val("").hide();
                console.error(data);          
                $("#success").text(data["msg"]).show();
            }
        ).fail(
            function(e) {
                $("#alert").show().text(e.responseText);
                console.error(e);                                             
            }
        );
    }
};
