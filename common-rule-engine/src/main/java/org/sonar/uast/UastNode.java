package org.sonar.uast;

import java.util.List;
import java.util.Optional;
import java.util.Set;
import java.util.function.Predicate;
import java.util.regex.Pattern;
import java.util.stream.Collectors;
import javax.annotation.Nullable;

public final class UastNode {

  public final Set<Kind> kinds;
  public final String nativeNode;
  @Nullable
  public final Token token;
  public final List<UastNode> children;

  public UastNode(Set<Kind> kinds, String nativeNode, @Nullable Token token, List<UastNode> children) {
    this.kinds = kinds;
    this.nativeNode = nativeNode;
    this.token = token;
    this.children = children;
  }

  public static class Token {

    private static final Pattern LINE_SPLITTER = Pattern.compile("\r\n|\n|\r");

    public final String value;
    /**
     * start at 1, line number of the first character of the token
     */
    public final int line;
    /**
     * start at 1, column number of the first character of the token
     */
    public final int column;
    /**
     * start at 1, line number of the last character of the token
     * if the token is on a single line, endLine() == line()
     */
    public final int endLine;
    /**
     * start at 1, column number of the last character of the token
     * if the token has only one character, endColumn() == column()
     * EOF token is empty, in this case endColumn() == column() - 1
     */
    public final int endColumn;

    public Token(int line, int column, String value) {
      if (line < 1 || column < 1) {
        throw new IllegalArgumentException("Invalid token location " + line + ":" + column);
      }
      this.line = line;
      this.column = column;
      this.value = value;
      if (value.indexOf('\n') == -1 && value.indexOf('\r') == -1) {
        this.endLine = line;
        this.endColumn = column + value.length() - 1;
      } else {
        String[] lines = LINE_SPLITTER.split(value, -1);
        this.endLine = line + lines.length - 1;
        this.endColumn = lines[lines.length - 1].length();
      }
    }

    @Override
    public String toString() {
      return "[" + line + ":" + column + " " + value + "]";
    }
  }

  public enum Kind implements Predicate<UastNode> {
    ASSIGNMENT,
    ASSIGNMENT_OPERATOR,
    ASSIGNMENT_TARGET,
    ASSIGNMENT_VALUE,
    BLOCK,
    CLASS,
    COMMENT,
    COMPILATION_UNIT,
    FUNCTION,
    IDENTIFIER,
    PARAMETER,
    STATEMENT,
    UNSUPPORTED,
    EOF
    ;

    @Override
    public boolean test(UastNode uastNode) {
      return uastNode.kinds.contains(this);
    }
  }

  public Optional<UastNode> getChild(Kind kind) {
    return children.stream().filter(kind).findAny();
  }

  public List<UastNode> getChildren(Kind... kinds) {
    return children.stream()
      .filter(child -> {
        for (Kind kind : kinds) {
          if (child.kinds.contains(kind)) {
            return true;
          }
        }
        return false;
      })
      .collect(Collectors.toList());
  }

  public UastNode firstToken() {
    if (token != null) {
      return this;
    }
    for (UastNode child : children) {
      UastNode uastNode = child.firstToken();
      if (uastNode != null) {
        return uastNode;
      }
    }
    return null;
  }

  @Override
  public String toString() {
    if (token != null) {
      return token.toString();
    }
    UastNode firstToken = firstToken();
    return kinds.toString() + (firstToken != null ? firstToken.toString() : "");
  }
}
