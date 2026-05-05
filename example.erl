-module(sample).
-export([sum/2]).

% Comentario en Erlang
sum(A, B) when A >= 0, B >= 0 ->
    Result = A + B,
    io:format("resultado=~p~n", [Result]),
    Result.
