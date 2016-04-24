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
    var name = $("#name").val();
    if (name == "") {
        $("#name-holder").removeClass("has-warning has-default").addClass("has-error");
        $("#alert").show().text("Please provide an email address.");
    } else {
        $.get(
                "/api/v0/register/?email="+name
             ).done(function(data) {
            $("#name").val("");
            console.log(data);          
        }).fail(function(e) {
                 $("#alert").show().text(e.statusText + ": " + e.responseText);
                 console.error(e);                                             
             });
    }
};
